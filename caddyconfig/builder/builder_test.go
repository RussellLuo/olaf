package builder

import (
	"reflect"
	"strings"
	"testing"

	"github.com/RussellLuo/olaf"
)

func TestFindAppliedPlugins(t *testing.T) {
	cases := []struct {
		name        string
		inPlugins   map[string]*olaf.Plugin
		inRoute     *olaf.Route
		wantPlugins []*olaf.Plugin
	}{
		{
			name: "service's goes before global's",
			inPlugins: map[string]*olaf.Plugin{
				"global_plugin_1": {
					Name: "global_plugin_1",
					Type: "request_body_var",
				},
				"service_1_plugin_1": {
					Name:        "service_1_plugin_1",
					Type:        "request_body_var",
					ServiceName: "service_1",
				},
			},
			inRoute: &olaf.Route{
				Name:        "route_1",
				ServiceName: "service_1",
			},
			wantPlugins: []*olaf.Plugin{
				{
					Name:        "service_1_plugin_1",
					Type:        "request_body_var",
					ServiceName: "service_1",
				},
			},
		},
		{
			name: "route's goes before service's",
			inPlugins: map[string]*olaf.Plugin{
				"service_1_plugin_1": {
					Name:        "service_1_plugin_1",
					Type:        "request_body_var",
					ServiceName: "service_1",
				},
				"route_1_plugin_1": {
					Name:      "route_1_plugin_1",
					Type:      "request_body_var",
					RouteName: "route_1",
				},
			},
			inRoute: &olaf.Route{
				Name:        "route_1",
				ServiceName: "service_1",
			},
			wantPlugins: []*olaf.Plugin{
				{
					Name:      "route_1_plugin_1",
					Type:      "request_body_var",
					RouteName: "route_1",
				},
			},
		},
		{
			name: "route-service's goes before route's",
			inPlugins: map[string]*olaf.Plugin{
				"route_1_plugin_1": {
					Name:      "route_1_plugin_1",
					Type:      "request_body_var",
					RouteName: "route_1",
				},
				"service_1_route_1_plugin_1": {
					Name:        "service_1_route_1_plugin_1",
					Type:        "request_body_var",
					RouteName:   "route_1",
					ServiceName: "service_1",
				},
			},
			inRoute: &olaf.Route{
				Name:        "route_1",
				ServiceName: "service_1",
			},
			wantPlugins: []*olaf.Plugin{
				{
					Name:        "service_1_route_1_plugin_1",
					Type:        "request_body_var",
					RouteName:   "route_1",
					ServiceName: "service_1",
				},
			},
		},
		{
			name: "complicated case",
			inPlugins: map[string]*olaf.Plugin{
				"global_plugin_1": {
					Name: "global_plugin_1",
					Type: "request_body_var",
				},
				"service_1_plugin_1": {
					Name:        "service_1_plugin_1",
					Type:        "rate_limit",
					OrderAfter:  "request_body_var",
					ServiceName: "service_1",
				},
				"route_1_plugin_1": {
					Name:       "route_1_plugin_1",
					Type:       "rate_limit",
					OrderAfter: "request_body_var",
					RouteName:  "route_1",
				},
				"service_1_route_1_plugin_1": {
					Name:        "service_1_route_1_plugin_1",
					Type:        "canary",
					OrderAfter:  "rate_limit",
					RouteName:   "route_1",
					ServiceName: "service_1",
				},
			},
			inRoute: &olaf.Route{
				Name:        "route_1",
				ServiceName: "service_1",
			},
			wantPlugins: []*olaf.Plugin{
				{
					Name: "global_plugin_1",
					Type: "request_body_var",
				},
				{
					Name:       "route_1_plugin_1",
					Type:       "rate_limit",
					OrderAfter: "request_body_var",
					RouteName:  "route_1",
				},
				{
					Name:        "service_1_route_1_plugin_1",
					Type:        "canary",
					OrderAfter:  "rate_limit",
					RouteName:   "route_1",
					ServiceName: "service_1",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			plugins, _ := findAppliedPlugins(c.inPlugins, c.inRoute)
			if !reflect.DeepEqual(plugins, c.wantPlugins) {
				t.Fatalf("Plugins: got (%+v), want (%+v)", plugins, c.wantPlugins)
			}
		})
	}
}

func TestSortPluginsByOrderAfter(t *testing.T) {
	cases := []struct {
		name           string
		inTypedPlugins map[string]*olaf.Plugin
		wantPlugins    []*olaf.Plugin
		wantErrStr     string
	}{
		{
			name: "one plugin",
			inTypedPlugins: map[string]*olaf.Plugin{
				"request_body_var": {
					Name: "plugin_1",
					Type: "request_body_var",
				},
			},
			wantPlugins: []*olaf.Plugin{
				{
					Name: "plugin_1",
					Type: "request_body_var",
				},
			},
		},
		{
			name: "multiple plugins",
			inTypedPlugins: map[string]*olaf.Plugin{
				"request_body_var": {
					Name: "plugin_1",
					Type: "request_body_var",
				},
				"rate_limit": {
					Name:       "plugin_2",
					Type:       "rate_limit",
					OrderAfter: "request_body_var",
				},
				"canary": {
					Name:       "plugin_3",
					Type:       "canary",
					OrderAfter: "rate_limit",
				},
			},
			wantPlugins: []*olaf.Plugin{
				{
					Name: "plugin_1",
					Type: "request_body_var",
				},
				{
					Name:       "plugin_2",
					Type:       "rate_limit",
					OrderAfter: "request_body_var",
				},
				{
					Name:       "plugin_3",
					Type:       "canary",
					OrderAfter: "rate_limit",
				},
			},
		},
		{
			name: "circular order dependency",
			inTypedPlugins: map[string]*olaf.Plugin{
				"request_body_var": {
					Name:       "plugin_1",
					Type:       "request_body_var",
					OrderAfter: "rate_limit",
				},
				"rate_limit": {
					Name:       "plugin_2",
					Type:       "rate_limit",
					OrderAfter: "request_body_var",
				},
			},
			// The error string here is just a prefix, since this test is
			// not stable due to the randomness of Go map.
			wantErrStr: "circular order dependency is detected for plugin",
		},
		{
			name: "plugin type not found",
			inTypedPlugins: map[string]*olaf.Plugin{
				"rate_limit": {
					Name:       "plugin_1",
					Type:       "rate_limit",
					OrderAfter: "request_body_var",
				},
				"canary": {
					Name:       "plugin_2",
					Type:       "canary",
					OrderAfter: "rate_limit",
				},
			},
			wantErrStr: `plugin type "request_body_var" (depended by plugin "plugin_1") not found`,
		},
		{
			name: "plugin unordered",
			inTypedPlugins: map[string]*olaf.Plugin{
				"request_body_var": {
					Name: "plugin_1",
					Type: "request_body_var",
				},
				"rate_limit": {
					Name: "plugin_2",
					Type: "rate_limit",
				},
				"canary": {
					Name:       "plugin_3",
					Type:       "canary",
					OrderAfter: "rate_limit",
				},
			},
			wantErrStr: `plugin "plugin_1" (of type "request_body_var") is unordered`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			plugins, err := sortPluginsByOrderAfter(c.inTypedPlugins)
			if err != nil && !strings.HasPrefix(err.Error(), c.wantErrStr) {
				t.Fatalf("ErrStr: got (%#v), want (%#v)", err.Error(), c.wantErrStr)
			}
			if !reflect.DeepEqual(plugins, c.wantPlugins) {
				t.Fatalf("Plugins: got (%+v), want (%+v)", plugins, c.wantPlugins)
			}
		})
	}
}

func TestPluginCanaryExpression(t *testing.T) {
	cases := []struct {
		name       string
		inPlugin   *olaf.Plugin
		inServices map[string]*olaf.Service
		wantMatch  map[string]interface{}
	}{
		{
			name: "canary per path",
			inPlugin: &olaf.Plugin{
				Type: olaf.PluginTypeCanary,
				Config: map[string]interface{}{
					"upstream":  "staging",
					"key":       "{path.0}",
					"whitelist": `$.startsWith("tid")`,
				},
			},
			inServices: map[string]*olaf.Service{
				"staging": {
					Name: "staging",
					URL:  "localhost:8080",
				},
			},
			wantMatch: map[string]interface{}{
				"expression": `{http.request.uri.path.0}.startsWith("tid")`,
			},
		},
		{
			name: "canary per query",
			inPlugin: &olaf.Plugin{
				Type: olaf.PluginTypeCanary,
				Config: map[string]interface{}{
					"upstream":  "staging",
					"key":       "{query.tid}",
					"type":      "int",
					"whitelist": "$ > 0 && $ <= 10",
				},
			},
			inServices: map[string]*olaf.Service{
				"staging": {
					Name: "staging",
					URL:  "localhost:8080",
				},
			},
			wantMatch: map[string]interface{}{
				"expression": "int({http.request.uri.query.tid}) > 0 && int({http.request.uri.query.tid}) <= 10",
			},
		},
		{
			name: "canary per body",
			inPlugin: &olaf.Plugin{
				Type: olaf.PluginTypeCanary,
				Config: map[string]interface{}{
					"upstream":  "staging",
					"key":       "{body.tid}",
					"type":      "int",
					"whitelist": "$ > 0 && $ <= 10",
				},
			},
			inServices: map[string]*olaf.Service{
				"staging": {
					Name: "staging",
					URL:  "localhost:8080",
				},
			},
			wantMatch: map[string]interface{}{
				"expression": "int({http.request.body.tid}) > 0 && int({http.request.body.tid}) <= 10",
			},
		},
		{
			name: "advanced matcher",
			inPlugin: &olaf.Plugin{
				Type: olaf.PluginTypeCanary,
				Config: map[string]interface{}{
					"upstream": "staging",
					"matcher": map[string]interface{}{
						"expression": `{http.request.uri.path.0}.startsWith("tid")`,
					},
				},
			},
			inServices: map[string]*olaf.Service{
				"staging": {
					Name: "staging",
					URL:  "localhost:8080",
				},
			},
			wantMatch: map[string]interface{}{
				"expression": `{http.request.uri.path.0}.startsWith("tid")`,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			routes := canaryReverseProxy(c.inPlugin, c.inServices)
			matchList := routes[0]["match"].([]map[string]interface{})
			gotMatch := matchList[0] // Get the first match.

			if !reflect.DeepEqual(gotMatch, c.wantMatch) {
				t.Fatalf("Match: got (%#v), want (%#v)", gotMatch, c.wantMatch)
			}
		})
	}
}
