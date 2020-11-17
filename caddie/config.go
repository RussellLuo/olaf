package caddie

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/RussellLuo/olaf/admin"
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

	for _, route := range data.Routes {
		routes = append(routes, map[string]interface{}{
			"match": []map[string][]string{
				{
					"method": route.Methods,
					"host":   route.Hosts,
					"path":   route.Paths,
				},
			},
			"handle": []map[string]interface{}{
				{
					"handler": "subroute",
					"routes":  buildSubRoutes(route, services, plugins),
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

func buildSubRoutes(r *admin.Route, s map[string]*admin.Service, p map[string]*admin.TenantCanaryPlugin) (routes []map[string]interface{}) {
	routes = []map[string]interface{}{
		{
			"handle": []map[string]string{
				{
					"handler":           "rewrite",
					"strip_path_prefix": r.StripPrefix,
				},
			},
		},
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

	for _, rr := range canaryReverseProxy(plugin) {
		routes = append(routes, rr)
	}

	routes = append(routes, reverseProxy(service.URL, ""))

	return
}

func canaryReverseProxy(p *admin.TenantCanaryPlugin) (routes []map[string]interface{}) {
	if p == nil {
		return
	}

	name := p.Config.TenantIDName
	idVar := fmt.Sprintf("int({http.request.uri.query.%s})", name)

	if p.Config.TenantIDLocation != "query" || name == "" {
		return
	}

	if len(p.Config.TenantIDList) > 0 {
		csv := strings.Replace(fmt.Sprint(p.Config.TenantIDList), " ", ",", -1)
		expr := fmt.Sprintf("%s in %s", idVar, csv)
		routes = append(routes, reverseProxy(p.Config.UpstreamURL, expr))
	}

	start := p.Config.TenantIDRange.Start
	end := p.Config.TenantIDRange.End
	if start != 0 || end != 0 {
		expr := fmt.Sprintf(
			"%s >=%d && %s <= %d",
			idVar, start,
			idVar, end)
		routes = append(routes, reverseProxy(p.Config.UpstreamURL, expr))
	}

	return
}

func reverseProxy(url, expr string) map[string]interface{} {
	route := map[string]interface{}{
		"handle": []map[string]interface{}{
			{
				"handler": "reverse_proxy",
				"upstreams": []map[string]string{
					{
						"dial": url,
					},
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
