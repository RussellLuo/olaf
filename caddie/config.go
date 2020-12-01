package caddie

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/RussellLuo/olaf/admin"
)

var (
	reRegexpPath = regexp.MustCompile(`~(\w+)?:\s*(.+)`)
)

func buildCaddyConfig(config HTTPConfig, data *Data) map[string]interface{} {
	routes := buildCaddyRoutes(data)

	return map[string]interface{}{
		"apps": map[string]interface{}{
			"http": map[string]interface{}{
				"http_port":  config.HTTPPort,
				"https_port": config.HTTPSPort,
				"servers":    buildServers(config.Servers, routes),
			},
		},
	}
}

func buildCaddyRoutes(data *Data) (routes []map[string]interface{}) {
	services := data.Services
	plugins := data.Plugins

	for _, r := range data.Routes {
		if services[r.ServiceName] == nil {
			log.Printf("service %q of route %q not found", r.ServiceName, r.Name)
			continue
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
		"handle": []map[string]string{
			{
				"handler": "static_response",
				"body":    "no content",
			},
		},
	})
	return
}

func buildRouteMatches(r *admin.Route) (matches []map[string]interface{}) {
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

func buildSubRoutes(r *admin.Route, s map[string]*admin.Service, p map[string]*admin.TenantCanaryPlugin) (routes []map[string]interface{}) {
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

	service := s[r.ServiceName]

	// Use the route-level plugin first, if any.
	plugin, ok := p[admin.PluginTenantCanary+"@"+r.Name]
	if !ok || !plugin.Enabled {
		// Then use the service-level plugin.
		plugin, ok = p[admin.PluginTenantCanary+"@"+service.Name]
		if !ok || !plugin.Enabled {
			plugin = nil
		}
	}

	canaryRoutes, canaryFieldInBody := canaryReverseProxy(plugin)
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

	routes = append(routes, reverseProxy(
		service.URL,
		service.DialTimeout,
		service.MaxRequests,
		"",
	))

	return
}

func canaryReverseProxy(p *admin.TenantCanaryPlugin) (routes []map[string]interface{}, canaryFieldInBody bool) {
	if p == nil {
		return
	}

	name := p.Config.TenantIDName
	if name == "" {
		return
	}

	var idVar string
	switch p.Config.TenantIDLocation {
	case "query":
		idVar = fmt.Sprintf("int({http.request.uri.query.%s})", name)
	case "body":
		idVar = fmt.Sprintf("int({http.request.body.%s})", name)
		canaryFieldInBody = true
	default:
		return
	}

	if len(p.Config.TenantIDList) > 0 {
		csv := strings.Replace(fmt.Sprint(p.Config.TenantIDList), " ", ",", -1)
		expr := fmt.Sprintf("%s in %s", idVar, csv)
		routes = append(routes, reverseProxy(
			p.Config.UpstreamURL,
			p.Config.UpstreamDialTimeout,
			p.Config.UpstreamMaxRequests,
			expr,
		))
	}

	start := p.Config.TenantIDRange.Start
	end := p.Config.TenantIDRange.End
	if start != 0 || end != 0 {
		expr := fmt.Sprintf(
			"%s >=%d && %s <= %d",
			idVar, start, idVar, end)
		routes = append(routes, reverseProxy(
			p.Config.UpstreamURL,
			p.Config.UpstreamDialTimeout,
			p.Config.UpstreamMaxRequests,
			expr,
		))
	}

	return
}

func reverseProxy(url, dialTimeout string, maxRequests int, expr string) map[string]interface{} {
	var timeout time.Duration
	if dialTimeout != "" {
		var err error
		timeout, err = time.ParseDuration(dialTimeout)
		if err != nil {
			log.Printf("parse dial_timeout err: %v\n", err)
		}
	}

	route := map[string]interface{}{
		"handle": []map[string]interface{}{
			{
				"handler": "reverse_proxy",
				"upstreams": []map[string]interface{}{
					buildUpstream(url, maxRequests),
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

func buildServers(addrs []string, routes []map[string]interface{}) map[string]interface{} {
	listenHosts := make(map[string][]string)
	for _, a := range addrs {
		s := strings.SplitN(a, ":", 2)
		host, listen := s[0], ":"+s[1]
		listenHosts[listen] = append(listenHosts[listen], host)
	}

	i := 0
	servers := make(map[string]interface{})

	for listen, hosts := range listenHosts {
		name := fmt.Sprintf("srv%d", i)
		i++

		servers[name] = map[string]interface{}{
			"listen": []string{listen},
			"routes": []map[string]interface{}{
				{
					"terminal": true,
					"match": []map[string]interface{}{
						{
							"host": hosts,
						},
					},
					"handle": []map[string]interface{}{
						{
							"handler": "subroute",
							"routes":  routes,
						},
					},
				},
			},
		}
	}

	return servers
}

func setCaddyConfig(config map[string]interface{}) error {
	u := &url.URL{
		Scheme: "http",
		Host:   "localhost:2019",
		Path:   "/load",
	}

	reqBody, err := json.Marshal(config)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusNoContent {
		msg, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(msg))
	}

	return nil
}
