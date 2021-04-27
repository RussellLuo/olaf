package builder

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/RussellLuo/olaf"
	"github.com/mitchellh/mapstructure"
)

var (
	reRegexpPath = regexp.MustCompile(`~(\w+)?:\s*(.+)`)

	reTCPAddressFormat = regexp.MustCompile(`^([^:]+)?(:\d+(-\d+)?)?$`)

	reCanaryKeyVar = regexp.MustCompile(`^\{(\w+)\.(.+)\}$`)
)

const (
	networkPrefixTCP  = "tcp/"
	networkPrefixUDP  = "udp/"
	networkPrefixUnix = "unix/"

	networkTCP  = "tcp"
	networkUnix = "unix"

	loggerName = "log0"
)

func Build(data *olaf.Data) (conf map[string]interface{}, err error) {
	defer func(errPtr *error) {
		if err := recover(); err != nil {
			if e, ok := err.(error); ok {
				*errPtr = e
			}
		}
	}(&err)

	routes := buildCaddyRoutes(data)

	conf = map[string]interface{}{
		"apps": map[string]interface{}{
			"http": map[string]interface{}{
				"http_port":  data.Server.HTTPPort,
				"https_port": data.Server.HTTPSPort,
				"servers": buildServers(
					data.Server.Listen, data.Server.EnableAutoHTTPS,
					data.Server.AccessLog.Disabled, routes,
				),
			},
		},
	}

	// Add the admin settings.
	conf["admin"] = buildAdminConfig(&data.Server.Admin)

	// Add the logging settings.
	conf["logging"] = buildLoggingConfig(data.Server.EnableDebug, &data.Server.DefaultLog, &data.Server.AccessLog)

	return
}

func buildCaddyRoutes(data *olaf.Data) (routes []map[string]interface{}) {
	services := data.Services
	plugins := data.Plugins

	// Build the routes from highest priority to lowest.
	// The route that has a higher priority will be matched earlier.
	for _, r := range sortRoutes(data.Routes) {
		if services[r.ServiceName] == nil {
			panic(fmt.Errorf("service %q of route %q not found", r.ServiceName, r.Name))
		}

		routes = append(routes, map[string]interface{}{
			"match": buildRouteMatches(r.Matcher),
			"handle": []map[string]interface{}{
				{
					"handler": "subroute",
					"routes":  buildSubRoutes(r, services, plugins),
				},
			},
		})
	}

	return
}

func buildStaticResponse(statusCode int, headers map[string][]string, body string, close bool) []map[string]interface{} {
	m := map[string]interface{}{
		"handler":     "static_response",
		"status_code": statusCode,
	}
	if len(headers) > 0 {
		m["headers"] = headers
	}
	if body != "" {
		m["body"] = body
	}
	if close {
		m["close"] = close
	}
	return []map[string]interface{}{m}
}

// sortRoutes sorts the given routes from highest priority to lowest.
func sortRoutes(r map[string]*olaf.Route) (routes []*olaf.Route) {
	for _, route := range r {
		routes = append(routes, route)
	}

	sort.SliceStable(routes, func(i, j int) bool {
		return routes[i].Priority > routes[j].Priority
	})

	return
}

func buildRouteMatches(matcher olaf.Matcher) (matches []map[string]interface{}) {
	// Differentiate regexp paths from normal paths.
	var normalPaths []string
	regexPaths := make(map[string]string)
	for _, p := range matcher.Paths {
		result := reRegexpPath.FindStringSubmatch(p)
		if len(result) > 0 {
			name, path := result[1], result[2]
			regexPaths[path] = name // We assume that name is globally unique if non-empty
		} else {
			normalPaths = append(normalPaths, p)
		}
	}

	buildMatch := func(m map[string]interface{}) map[string]interface{} {
		if matcher.Protocol != "" {
			m["protocol"] = matcher.Protocol
		}
		if len(matcher.Methods) > 0 {
			m["method"] = matcher.Methods
		}
		if len(matcher.Hosts) > 0 {
			m["host"] = matcher.Hosts
		}
		if len(matcher.Headers) > 0 {
			m["header"] = matcher.Headers
		}
		return m
	}

	// Build a match for normal paths.
	if len(normalPaths) > 0 {
		matches = append(matches, buildMatch(map[string]interface{}{
			"path": normalPaths,
		}))
	}

	// Build a match for regexp paths.
	for p, n := range regexPaths {
		matches = append(matches, buildMatch(map[string]interface{}{
			"path_regexp": map[string]string{
				"name":    n,
				"pattern": p,
			},
		}))
	}

	return
}

func buildSubRoutes(r *olaf.Route, services map[string]*olaf.Service, plugins map[string]*olaf.Plugin) (routes []map[string]interface{}) {
	if r.Response != nil {
		// This is a STATIC route, any other PROXY-related attributes will be ignored.
		routes = append(routes, map[string]interface{}{
			"handle": buildStaticResponse(r.Response.StatusCode, r.Response.Headers, r.Response.Body, r.Response.Close),
		})
		return
	}

	// This is a PROXY route.

	routes = append(routes, manipulateURI(r.URI, nil)...)

	appliedPlugins, err := findAppliedPlugins(plugins, r)
	if err != nil {
		panic(err)
	}

	// Build routes from plugins.
	for _, p := range appliedPlugins {
		switch p.Type {
		case olaf.PluginTypeCanary: // For the built-in canary plugin.
			canaryRoutes := canaryReverseProxy(p, services)
			routes = append(routes, canaryRoutes...)
		default: // For other plugins (usually third-party Caddy extensions).
			routes = append(routes, buildPluginRoute(p))
		}
	}

	// Normal reverse-proxy routes must come after canary reverse-proxy routes.
	service := services[r.ServiceName]
	routes = append(routes, reverseProxy(service, nil))
	return
}

func manipulateURI(uri olaf.URI, matcher map[string]interface{}) (routes []map[string]interface{}) {
	if uri.StripPrefix != "" || uri.StripSuffix != "" {
		stripHandler := map[string]string{
			"handler": "rewrite",
		}
		if uri.StripPrefix != "" {
			stripHandler["strip_path_prefix"] = uri.StripPrefix
		}
		if uri.StripSuffix != "" {
			stripHandler["strip_path_suffix"] = uri.StripSuffix
		}
		routes = append(routes, addRouteMatcher(
			map[string]interface{}{
				"handle": []map[string]string{stripHandler},
			},
			matcher,
		))
	}

	targetPath := uri.TargetPath
	if targetPath == "" && uri.AddPrefix != "" {
		// TODO: Deprecate AddPrefix
		// Convert AddPrefix to TargetPath for compatibility.
		targetPath = uri.AddPrefix + "$"
	}
	if targetPath != "" {
		// `$` is a placeholder for the path component of the request URI,
		// and it should occur at most once.
		targetPath = strings.Replace(targetPath, "$", "{http.request.uri.path}", 1)
		routes = append(routes, addRouteMatcher(
			map[string]interface{}{
				"handle": []map[string]string{
					{
						"handler": "rewrite",
						"uri":     targetPath,
					},
				},
			},
			matcher,
		))
	}

	return
}

func addRouteMatcher(route, matcher map[string]interface{}) map[string]interface{} {
	if len(route) > 0 && len(matcher) > 0 {
		route["match"] = []map[string]interface{}{matcher}
	}
	return route
}

// findAppliedPlugins finds the plugins that have been applied to the given route.
func findAppliedPlugins(plugins map[string]*olaf.Plugin, r *olaf.Route) ([]*olaf.Plugin, error) {
	routeServicePlugins := make(map[string][]*olaf.Plugin)
	routePlugins := make(map[string][]*olaf.Plugin)
	servicePlugins := make(map[string][]*olaf.Plugin)
	var globalPlugins []*olaf.Plugin

	for _, p := range plugins {
		if p.Disabled {
			continue
		}

		switch {
		case p.RouteName != "" && p.ServiceName != "":
			routeServicePlugins[p.RouteName] = append(routeServicePlugins[p.RouteName], p)
		case p.RouteName != "":
			routePlugins[p.RouteName] = append(routePlugins[p.RouteName], p)
		case p.ServiceName != "":
			servicePlugins[p.ServiceName] = append(servicePlugins[p.ServiceName], p)
		default:
			globalPlugins = append(globalPlugins, p)
		}
	}

	// A plugin (of the same type) will always be run once and only once per request.
	// The plugin precedence follows https://docs.konghq.com/2.0.x/admin-api/#precedence
	typedPlugins := make(map[string]*olaf.Plugin)
	addPlugin := func(p *olaf.Plugin) {
		if _, ok := typedPlugins[p.Type]; !ok {
			typedPlugins[p.Type] = p
		}
	}
	for _, p := range routeServicePlugins[r.Name] {
		addPlugin(p)
	}
	for _, p := range routePlugins[r.Name] {
		addPlugin(p)
	}
	for _, p := range servicePlugins[r.ServiceName] {
		addPlugin(p)
	}
	for _, p := range globalPlugins {
		addPlugin(p)
	}

	return sortPluginsByOrderAfter(typedPlugins)
}

// sortPluginsByOrderAfter sorts the plugins according to the `OrderAfter` field.
func sortPluginsByOrderAfter(typedPlugins map[string]*olaf.Plugin) (plugins []*olaf.Plugin, err error) {
	// If there is only one plugin, return it immediately.
	if len(typedPlugins) == 1 {
		for _, p := range typedPlugins {
			plugins = append(plugins, p)
		}
		return
	}

	processed := make(map[string]bool)
	emptyOrderAfterPlugins := make(map[string]*olaf.Plugin)

	for _, p := range typedPlugins {
		if p.OrderAfter == "" {
			emptyOrderAfterPlugins[p.Type] = p
			continue
		}

		// Build the stack per the order dependency.
		pendingTypes := make(map[string]bool)
		var stack []*olaf.Plugin

		for {
			if _, ok := processed[p.Type]; ok {
				if _, ok := pendingTypes[p.Type]; ok {
					return nil, fmt.Errorf("circular order dependency is detected for plugin %q (of type %q)", p.Name, p.Type)
				}
				break
			}
			processed[p.Type] = true
			pendingTypes[p.Type] = true

			stack = append(stack, p)

			if p.OrderAfter == "" {
				break
			}

			afterP, ok := typedPlugins[p.OrderAfter]
			if !ok || afterP == nil {
				return nil, fmt.Errorf("plugin type %q (depended by plugin %q) not found", p.OrderAfter, p.Name)
			}
			p = afterP
		}

		// Append the plugins in the reverse order.
		for i := len(stack); i > 0; i-- {
			p := stack[i-1]
			plugins = append(plugins, p)
		}
	}

	for t, p := range emptyOrderAfterPlugins {
		if _, ok := processed[t]; !ok {
			return nil, fmt.Errorf("plugin %q (of type %q) is unordered", p.Name, p.Type)
		}
	}

	return
}

func buildPluginRoute(p *olaf.Plugin) map[string]interface{} {
	handle := map[string]interface{}{
		"handler": p.Type,
	}
	for k, v := range p.Config {
		handle[k] = v
	}
	return map[string]interface{}{
		"handle": []map[string]interface{}{handle},
	}
}

func canaryReverseProxy(p *olaf.Plugin, services map[string]*olaf.Service) (routes []map[string]interface{}) {
	if p == nil || p.Type != olaf.PluginTypeCanary {
		return
	}

	config := new(olaf.PluginCanaryConfig)
	if err := mapstructure.Decode(p.Config, config); err != nil {
		panic(fmt.Errorf("config of plugin %q cannot be decoded into olaf.PluginCanaryConfig", p.Name))
	}

	s := services[config.UpstreamServiceName]
	if s == nil {
		panic(fmt.Errorf("upstream service %q of plugin %q not found", config.UpstreamServiceName, p.Name))
	}

	// If the advanced matcher is provided, use it instead.
	if len(config.Matcher) > 0 {
		if config.KeyName != "" || config.KeyType != "" || config.Whitelist != "" {
			panic(fmt.Errorf("invalid config of plugin %q: `matcher` and (`key`, `type`, `whitelist`) are mutually exclusive", p.Name))
		}

		routes = append(routes, manipulateURI(config.URI, config.Matcher)...)
		routes = append(routes, reverseProxy(s, config.Matcher))
		return
	}

	// Use the simple matcher `expression`.

	keyVar, err := parseVar(config.KeyName, p)
	if err != nil {
		panic(err)
	}

	// Do the type conversion if specified.
	if config.KeyType != "" {
		keyVar = fmt.Sprintf("%s(%s)", config.KeyType, keyVar)
	}

	if config.Whitelist == "" {
		panic(fmt.Errorf("whitelist of canary plugin %q is empty", p.Name))
	}
	expr := strings.ReplaceAll(config.Whitelist, "$", keyVar)

	matcher := map[string]interface{}{
		"expression": expr,
	}
	routes = append(routes, manipulateURI(config.URI, matcher)...)
	routes = append(routes, reverseProxy(s, matcher))

	return
}

// parseVar transforms shorthand variables into Caddy-style placeholders.
//
// Examples for shorthand variables:
//
//     {path.<var>}
//     {query.<var>}
//     {header.<VAR>}
//     {cookie.<var>}
//     {body.<var>}
//
func parseVar(s string, p *olaf.Plugin) (v string, err error) {
	result := reCanaryKeyVar.FindStringSubmatch(s)
	if len(result) != 3 {
		return "", fmt.Errorf("invalid key %q for canary plugin %q", s, p.Name)
	}
	location, name := result[1], result[2]

	switch location {
	case "path":
		v = fmt.Sprintf("{http.request.uri.path.%s}", name)
	case "query":
		v = fmt.Sprintf("{http.request.uri.query.%s}", name)
	case "header":
		v = fmt.Sprintf("{http.request.header.%s}", name)
	case "cookie":
		v = fmt.Sprintf("{http.request.cookie.%s}", name)
	case "body":
		v = fmt.Sprintf("{http.request.body.%s}", name)
	default:
		err = fmt.Errorf("unrecognized key %q for canary plugin %q", s, p.Name)
	}

	return
}

func reverseProxy(s *olaf.Service, matcher map[string]interface{}) map[string]interface{} {
	var timeout time.Duration
	if s.DialTimeout != "" {
		var err error
		timeout, err = time.ParseDuration(s.DialTimeout)
		if err != nil {
			panic(fmt.Errorf("failed to parse dial_timeout of service %q: %v", s.Name, err))
		}
	}

	handle := map[string]interface{}{
		"handler": "reverse_proxy",
		"upstreams": []map[string]interface{}{
			buildUpstream(s.URL, s.MaxRequests),
		},
		"transport": map[string]interface{}{
			"protocol":     "http",
			"dial_timeout": timeout,
		},
	}

	// Manipulate headers if requested.
	headers := make(map[string]interface{})
	if s.HeaderUp != nil {
		headers["request"] = manipulateHeader(s.HeaderUp)
	}
	if s.HeaderDown != nil {
		headers["response"] = manipulateHeader(s.HeaderDown)
	}
	if len(headers) > 0 {
		handle["headers"] = headers
	}

	route := map[string]interface{}{
		"handle": []map[string]interface{}{handle},
	}

	// Add possible matching rules.
	if len(matcher) > 0 {
		route["match"] = []map[string]interface{}{matcher}
	}

	return route
}

func manipulateHeader(h *olaf.HeaderOps) map[string]interface{} {
	m := make(map[string]interface{})
	if len(h.Set) > 0 {
		m["set"] = h.Set
	}
	if len(h.Add) > 0 {
		m["add"] = h.Add
	}
	if len(h.Delete) > 0 {
		m["delete"] = h.Delete
	}
	return m
}

func buildUpstream(url string, maxRequests int) map[string]interface{} {
	// Validate the conventional format of url.
	na, err := newNetAddr(url)
	if err != nil {
		panic(err)
	}
	// Special validation logic for TCP addresses to dial.
	// See https://caddyserver.com/docs/json/apps/http/servers/routes/handle/reverse_proxy/upstreams/dial#docs
	if na.Network == networkTCP {
		// A TCP address must have a host and a port.
		s := strings.SplitN(na.Address, ":", 2)
		port := s[1]

		// TCP address to dial can not use port ranges
		if strings.Contains(port, "-") {
			panic(fmt.Errorf("invalid TCP address: %q", url))
		}
	}

	m := map[string]interface{}{
		"dial": na.Address,
	}

	if maxRequests > 0 {
		m["max_requests"] = maxRequests
	}

	return m
}

// TODO: Use caddy.NetworkAddress directly.
type netAddr struct {
	Network string
	Address string
}

func newNetAddr(s string) (na netAddr, err error) {
	// See https://caddyserver.com/docs/conventions#network-addresses
	switch {
	case strings.HasPrefix(s, networkPrefixTCP):
		na.Network = networkTCP
		na.Address = addDefaultPort(strings.TrimPrefix(s, networkPrefixTCP))

		if na.Address == "" || !reTCPAddressFormat.MatchString(na.Address) {
			return na, fmt.Errorf("invalid TCP address: %q", s)
		}
	case strings.HasPrefix(s, networkPrefixUDP):
		return na, fmt.Errorf("unsupported UDP address: %q", s)
	case strings.HasPrefix(s, networkPrefixUnix):
		na.Network = networkUnix
		na.Address = s // Preserve the complete address.

		if !strings.HasPrefix(na.Address, networkPrefixUnix+"/") {
			return na, fmt.Errorf("invalid Unix address: %q", s)
		}
	default: // tcp
		na.Network = networkTCP
		na.Address = addDefaultPort(s)

		if na.Address == "" || !reTCPAddressFormat.MatchString(na.Address) {
			return na, fmt.Errorf("invalid TCP address: %q", s)
		}
	}

	return
}

func addDefaultPort(addr string) string {
	if addr == "" || strings.Contains(addr, ":") {
		// Return addr as is.
		return addr
	}

	// Add a default port if not specified.
	return addr + ":80"
}

func buildServers(addrs []string, enableAutoHTTPS, disableAccessLog bool, routes []map[string]interface{}) map[string]interface{} {
	listenHosts := make(map[string][]string)
	for _, a := range addrs {
		na, err := newNetAddr(a)
		if err != nil {
			panic(err)
		}

		switch na.Network {
		case networkTCP:
			// A TCP address must have a host and a port.
			s := strings.SplitN(na.Address, ":", 2)
			host, port := s[0], s[1]

			listen := ":" + port
			listenHosts[listen] = append(listenHosts[listen], host)
		case networkUnix:
			// Unix domain socket has no host.
			listenHosts[na.Address] = []string(nil)
		}
	}

	buildRoute := func(hosts []string) map[string]interface{} {
		r := map[string]interface{}{
			"handle": []map[string]interface{}{
				{
					"handler": "subroute",
					"routes":  routes,
				},
			},
		}

		matchAnyHost := false
		for _, h := range hosts {
			if h == "" {
				// Empty host indicates any host will match.
				matchAnyHost = true
				break
			}
		}

		if !matchAnyHost && len(hosts) > 0 {
			r["match"] = []map[string]interface{}{
				{
					"host": hosts,
				},
			}
			r["terminal"] = true
		}

		return r
	}

	i := 0
	servers := make(map[string]interface{})

	for listen, hosts := range listenHosts {
		name := fmt.Sprintf("srv%d", i)
		i++

		conf := map[string]interface{}{
			"automatic_https": map[string]interface{}{
				"disable": !enableAutoHTTPS,
			},
			"listen": []string{listen},
			"routes": []map[string]interface{}{
				buildRoute(hosts),
			},
		}
		// Add the logging settings if access-log is enabled.
		if !disableAccessLog {
			conf["logs"] = map[string]string{
				"default_logger_name": loggerName,
			}
		}

		servers[name] = conf
	}

	return servers
}

func buildAdminConfig(admin *olaf.Admin) map[string]interface{} {
	m := map[string]interface{}{
		"listen": admin.Listen,
		"config": map[string]bool{
			"persist": !admin.Nonpersistent,
		},
	}

	if admin.Disabled {
		m["disabled"] = admin.Disabled
	}
	if admin.EnforceOrigin {
		m["enforce_origin"] = admin.EnforceOrigin
	}
	if len(admin.Origins) > 0 {
		m["origins"] = admin.Origins
	}

	return m
}

func buildLoggingConfig(enableDebug bool, defaultLog *olaf.CaddyLog, accessLog *olaf.AccessLog) map[string]interface{} {
	level := defaultLog.Level
	if enableDebug {
		level = "DEBUG"
	}
	defaultLoggerConfig := map[string]interface{}{
		"writer": buildLogWriter(&defaultLog.Output),
		"level":  level,
	}
	logs := map[string]interface{}{
		"default": defaultLoggerConfig,
	}

	// if access-log is enabled.
	if !accessLog.Disabled {
		accessLoggerName := "http.log.access." + loggerName
		defaultLoggerConfig["exclude"] = []string{accessLoggerName}
		logs[loggerName] = map[string]interface{}{
			"include": []string{accessLoggerName},
			"writer":  buildLogWriter(&accessLog.Output),
			"level":   accessLog.Level,
		}
	}

	return map[string]interface{}{
		"logs": logs,
	}
}

func buildLogWriter(output *olaf.LogOutput) map[string]interface{} {
	writer := map[string]interface{}{
		"output": output.Output,
	}
	if output.Output == "file" {
		if output.Filename != "" {
			writer["filename"] = output.Filename
		}
		if output.RollDisabled {
			writer["roll"] = false
		}
		if output.RollSizeMB > 0 {
			writer["roll_size_mb"] = output.RollSizeMB
		}
		if output.RollKeep > 0 {
			writer["roll_keep"] = output.RollKeep
		}
		if output.RollKeepDays > 0 {
			writer["roll_keep_days"] = output.RollKeepDays
		}
	}
	return writer
}
