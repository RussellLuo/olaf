package builder

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/RussellLuo/olaf"
)

var (
	reRegexpPath = regexp.MustCompile(`~(\w+)?:\s*(.+)`)

	errUnsupportedNetwork = errors.New("unsupported network address")
)

const (
	networkPrefixTCP  = "tcp/"
	networkPrefixUDP  = "udp/"
	networkPrefixUnix = "unix/"

	networkTCP  = "tcp"
	networkUnix = "unix"
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

	// Add the logging settings if access-log is enabled.
	if !data.Server.DisableAccessLog {
		conf["logging"] = buildLoggingConfig()
	}

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
			"match": buildRouteMatches(r),
			"handle": []map[string]interface{}{
				{
					"handler": "subroute",
					"routes":  buildSubRoutes(r, services, plugins),
				},
			},
		})
	}

	routes = append(routes, map[string]interface{}{
		// Respond 404 for all unmatched routes.
		"handle": []map[string]interface{}{
			{
				"handler":     "static_response",
				"status_code": 404,
			},
		},
	})
	return
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

func buildRouteMatches(r *olaf.Route) (matches []map[string]interface{}) {
	// Differentiate regexp paths from normal paths
	var paths []string
	rePaths := make(map[string]string)
	for _, p := range r.Paths {
		result := reRegexpPath.FindStringSubmatch(p)
		if len(result) > 0 {
			name, path := result[1], result[2]
			rePaths[path] = name // We assume that name is globally unique if non-empty
		} else {
			paths = append(paths, p)
		}
	}

	// Build a match for normal paths
	if len(paths) > 0 {
		m := map[string]interface{}{
			"path": paths,
		}
		if len(r.Methods) > 0 {
			m["method"] = r.Methods
		}
		if len(r.Hosts) > 0 {
			m["host"] = r.Hosts
		}
		matches = append(matches, m)
	}

	// Build matches for regexp paths
	for p, n := range rePaths {
		m := map[string]interface{}{
			"path_regexp": map[string]string{
				"name":    n,
				"pattern": p,
			},
		}
		if len(r.Methods) > 0 {
			m["method"] = r.Methods
		}
		if len(r.Hosts) > 0 {
			m["host"] = r.Hosts
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

	for _, rr := range canaryRoutes {
		routes = append(routes, rr)
	}

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
		return
	}

	var idVar string
	switch p.Config.TenantIDLocation {
	case "path":
		idVar = fmt.Sprintf("int({http.request.uri.path.%s})", name)
	case "query":
		idVar = fmt.Sprintf("int({http.request.uri.query.%s})", name)
	case "header":
		idVar = fmt.Sprintf("int({http.request.header.%s})", name)
	case "cookie":
		idVar = fmt.Sprintf("int({http.request.cookie.%s})", name)
	case "body":
		idVar = fmt.Sprintf("int({http.request.body.%s})", name)
		canaryFieldInBody = true
	default:
		return
	}

	if p.Config.TenantIDWhitelist != "" {
		expr := strings.ReplaceAll(p.Config.TenantIDWhitelist, "$", idVar)
		routes = append(routes, reverseProxy(s, expr))
	}

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
	case strings.HasPrefix(s, networkPrefixUDP):
		return na, errUnsupportedNetwork
	case strings.HasPrefix(s, networkPrefixUnix):
		na.Network = networkUnix
		na.Address = s // Preserve the complete address.
	default: // tcp
		na.Network = networkTCP
		na.Address = s
	}

	return
}

func buildServers(addrs []string, enableAutoHTTPS, disableAccessLog bool, routes []map[string]interface{}) map[string]interface{} {
	listenHosts := make(map[string][]string)
	for _, a := range addrs {
		na, err := newNetAddr(a)
		if err != nil {
			panic(fmt.Errorf("unsupported network address: %q", a))
		}

		switch na.Network {
		case networkTCP:
			s := strings.SplitN(na.Address, ":", 2)
			host, listen := s[0], ":"+s[1]
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
				"default_logger_name": "log0",
			}
		}

		servers[name] = conf
	}

	return servers
}

func buildLoggingConfig() map[string]interface{} {
	return map[string]interface{}{
		"logs": map[string]interface{}{
			"default": map[string]interface{}{
				"exclude": []string{
					"http.log.access.log0",
				},
			},
			"log0": map[string]interface{}{
				"include": []string{
					"http.log.access.log0",
				},
				"writer": map[string]string{
					"output": "stdout",
				},
			},
		},
	}
}
