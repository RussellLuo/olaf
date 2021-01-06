package builder

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/RussellLuo/olaf"
)

var (
	reRegexpPath = regexp.MustCompile(`~(\w+)?:\s*(.+)`)

	reTCPAddressFormat = regexp.MustCompile(`^([^:]+)?(:\d+(-\d+)?)?$`)
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
					data.Server.DisableAccessLog, routes,
				),
			},
		},
	}

	// Add the logging settings.
	conf["logging"] = buildLoggingConfig(data.Server.DisableAccessLog, data.Server.EnableDebug)

	return
}

func buildCaddyRoutes(data *olaf.Data) (routes []map[string]interface{}) {
	services := data.Services
	plugins := data.Plugins

	// Build the routes for static before-responses, which will be matched
	// before all the services' routes.
	for _, r := range data.Server.BeforeResponses {
		routes = append(routes, map[string]interface{}{
			"match":  buildRouteMatches(r.Methods, r.Hosts, r.Paths),
			"handle": buildStaticResponse(r.StatusCode, r.Headers, r.Body, r.Close),
		})
	}

	// Build the routes from highest priority to lowest.
	// The route that has a higher priority will be matched earlier.
	for _, r := range sortRoutes(data.Routes) {
		if services[r.ServiceName] == nil {
			panic(fmt.Errorf("service %q of route %q not found", r.ServiceName, r.Name))
		}

		routes = append(routes, map[string]interface{}{
			"match": buildRouteMatches(r.Methods, r.Hosts, r.Paths),
			"handle": []map[string]interface{}{
				{
					"handler": "subroute",
					"routes":  buildSubRoutes(r, services, plugins),
				},
			},
		})
	}

	// Build the routes for static after-responses, which will be matched
	// after all the services' routes.
	for _, r := range data.Server.AfterResponses {
		routes = append(routes, map[string]interface{}{
			"match":  buildRouteMatches(r.Methods, r.Hosts, r.Paths),
			"handle": buildStaticResponse(r.StatusCode, r.Headers, r.Body, r.Close),
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

func buildRouteMatches(methods, hosts, paths []string) (matches []map[string]interface{}) {
	// Differentiate regexp paths from normal paths.
	var normalPaths []string
	regexPaths := make(map[string]string)
	for _, p := range paths {
		result := reRegexpPath.FindStringSubmatch(p)
		if len(result) > 0 {
			name, path := result[1], result[2]
			regexPaths[path] = name // We assume that name is globally unique if non-empty
		} else {
			normalPaths = append(normalPaths, p)
		}
	}

	// Build a match for normal paths.
	if len(normalPaths) > 0 {
		m := map[string]interface{}{
			"path": normalPaths,
		}
		if len(methods) > 0 {
			m["method"] = methods
		}
		if len(hosts) > 0 {
			m["host"] = hosts
		}
		matches = append(matches, m)
	}

	// Build matches for regexp paths.
	for p, n := range regexPaths {
		m := map[string]interface{}{
			"path_regexp": map[string]string{
				"name":    n,
				"pattern": p,
			},
		}
		if len(methods) > 0 {
			m["method"] = methods
		}
		if len(hosts) > 0 {
			m["host"] = hosts
		}
		matches = append(matches, m)
	}

	return
}

func buildSubRoutes(r *olaf.Route, services map[string]*olaf.Service, p map[string]*olaf.TenantCanaryPlugin) (routes []map[string]interface{}) {
	if r.StripPrefix != "" {
		routes = append(routes, map[string]interface{}{
			"handle": []map[string]string{
				{
					"handler":           "rewrite",
					"strip_path_prefix": r.StripPrefix,
				},
			},
		})
	}

	if r.AddPrefix != "" {
		routes = append(routes, map[string]interface{}{
			"handle": []map[string]string{
				{
					"handler": "rewrite",
					"uri":     r.AddPrefix + "{http.request.uri}",
				},
			},
		})
	}

	service := services[r.ServiceName]

	var plugin *olaf.TenantCanaryPlugin
	plugins := findAppliedPlugins(p, r)
	if len(plugins) > 0 {
		// For simplicity, only use the first plugin now.
		plugin = plugins[0]
	}

	canaryRoutes, canaryFieldInBody := canaryReverseProxy(plugin, services)
	if canaryFieldInBody {
		routes = append(routes, map[string]interface{}{
			"handle": []map[string]string{
				{
					"handler": "request_body_var",
				},
			},
		})
	}

	// Canary reverse-proxy routes must come before normal reverse-proxy routes.
	routes = append(routes, canaryRoutes...)

	routes = append(routes, reverseProxy(service, ""))
	return
}

// findAppliedPlugins finds the plugins that have been applied to the given route.
func findAppliedPlugins(ps map[string]*olaf.TenantCanaryPlugin, r *olaf.Route) []*olaf.TenantCanaryPlugin {
	routeServicePlugins := make(map[string][]*olaf.TenantCanaryPlugin)
	routePlugins := make(map[string][]*olaf.TenantCanaryPlugin)
	servicePlugins := make(map[string][]*olaf.TenantCanaryPlugin)
	var globalPlugins []*olaf.TenantCanaryPlugin

	for _, p := range ps {
		if !p.Enabled {
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

	// The plugin precedence follows https://docs.konghq.com/2.0.x/admin-api/#precedence
	plugins, ok := routeServicePlugins[r.Name]
	if !ok {
		plugins, ok = routePlugins[r.Name]
		if !ok {
			plugins, ok = servicePlugins[r.ServiceName]
			if !ok && len(globalPlugins) > 0 {
				plugins = globalPlugins
			}
		}
	}

	return plugins
}

func canaryReverseProxy(p *olaf.TenantCanaryPlugin, services map[string]*olaf.Service) (routes []map[string]interface{}, canaryFieldInBody bool) {
	if p == nil {
		return
	}

	s := services[p.Config.UpstreamServiceName]
	if s == nil {
		panic(fmt.Errorf("upstream service %q of plugin %q not found", p.Config.UpstreamServiceName, p.Name))
	}

	name := p.Config.TenantIDName
	if name == "" {
		panic(fmt.Errorf("tenant-id name of plugin %q is empty", p.Name))
	}

	var idVar string
	switch p.Config.TenantIDLocation {
	case "path":
		idVar = fmt.Sprintf("{http.request.uri.path.%s}", name)
	case "query":
		idVar = fmt.Sprintf("{http.request.uri.query.%s}", name)
	case "header":
		idVar = fmt.Sprintf("{http.request.header.%s}", name)
	case "cookie":
		idVar = fmt.Sprintf("{http.request.cookie.%s}", name)
	case "body":
		idVar = fmt.Sprintf("{http.request.body.%s}", name)
		canaryFieldInBody = true
	default:
		panic(fmt.Errorf("tenant-id location %q of plugin %q is invalid", p.Config.TenantIDLocation, p.Name))
	}

	// Do the type conversion if specified.
	if p.Config.TenantIDType != "" {
		idVar = fmt.Sprintf("%s(%s)", p.Config.TenantIDType, idVar)
	}

	if p.Config.TenantIDWhitelist == "" {
		panic(fmt.Errorf("tenant-id whitelist of plugin %q is empty", p.Name))
	}
	expr := strings.ReplaceAll(p.Config.TenantIDWhitelist, "$", idVar)
	routes = append(routes, reverseProxy(s, expr))

	return
}

func reverseProxy(s *olaf.Service, expr string) map[string]interface{} {
	var timeout time.Duration
	if s.DialTimeout != "" {
		var err error
		timeout, err = time.ParseDuration(s.DialTimeout)
		if err != nil {
			panic(fmt.Errorf("failed to parse dial_timeout of service %q: %v", s.Name, err))
		}
	}

	route := map[string]interface{}{
		"handle": []map[string]interface{}{
			{
				"handler": "reverse_proxy",
				"upstreams": []map[string]interface{}{
					buildUpstream(s.URL, s.MaxRequests),
				},
				"transport": map[string]interface{}{
					"protocol":     "http",
					"dial_timeout": timeout,
				},
			},
		},
	}

	if expr != "" {
		route["match"] = []map[string]string{
			{
				"expression": expr,
			},
		}
	}

	return route
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
		s := strings.SplitN(na.Address, ":", 2)
		if len(s) != 2 { // TCP address to dial must have a host and a port
			panic(fmt.Errorf("invalid TCP address: %q", url))
		}
		if strings.Contains(s[1], "-") { // TCP address to dial can not use port ranges
			panic(fmt.Errorf("invalid TCP address: %q", url))
		}
	}

	m := map[string]interface{}{
		"dial": url,
	}

	if maxRequests > 0 {
		m["max_requests"] = maxRequests
	}

	return m
}

type netAddr struct {
	Network string
	Address string
}

func newNetAddr(s string) (na netAddr, err error) {
	// See https://caddyserver.com/docs/conventions#network-addresses
	switch {
	case strings.HasPrefix(s, networkPrefixTCP):
		na.Network = networkTCP
		na.Address = strings.TrimPrefix(s, networkPrefixTCP)

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
		na.Address = s

		if na.Address == "" || !reTCPAddressFormat.MatchString(na.Address) {
			return na, fmt.Errorf("invalid TCP address: %q", s)
		}
	}

	return
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
			s := strings.SplitN(na.Address, ":", 2)
			host := s[0]

			listen := ":80"
			if len(s) == 2 {
				listen = ":" + s[1]
			}

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

func buildLoggingConfig(disableAccessLog, enableDebug bool) map[string]interface{} {
	level := "INFO"
	if enableDebug {
		level = "DEBUG"
	}
	defaultLog := map[string]interface{}{
		"level": level,
	}
	logs := map[string]interface{}{
		"default": defaultLog,
	}

	// if access-log is enabled.
	if !disableAccessLog {
		accessLoggerName := "http.log.access." + loggerName
		defaultLog["exclude"] = []string{accessLoggerName}
		logs[loggerName] = map[string]interface{}{
			"include": []string{accessLoggerName},
			"writer": map[string]string{
				"output": "stdout",
			},
		}
	}

	return map[string]interface{}{
		"logs": logs,
	}
}
