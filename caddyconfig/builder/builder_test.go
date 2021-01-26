package builder

import (
	"testing"

	"github.com/RussellLuo/olaf"
)

func TestPluginCanaryExpression(t *testing.T) {
	cases := []struct {
		inPlugin              *olaf.Plugin
		inServices            map[string]*olaf.Service
		wantMatchExpression   string
		wantCanaryFieldInBody bool
	}{
		{
			inPlugin: &olaf.Plugin{
				Type: olaf.PluginTypeCanary,
				Config: map[string]interface{}{
					"upstream":  "staging",
					"key":       "path.0",
					"whitelist": `$.startsWith("tid")`,
				},
			},
			inServices: map[string]*olaf.Service{
				"staging": {
					Name: "staging",
					URL:  "localhost:8080",
				},
			},
			wantMatchExpression: `{http.request.uri.path.0}.startsWith("tid")`,
		},
		{
			inPlugin: &olaf.Plugin{
				Type: olaf.PluginTypeCanary,
				Config: map[string]interface{}{
					"upstream":  "staging",
					"key":       "query.tid",
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
			wantMatchExpression: "int({http.request.uri.query.tid}) > 0 && int({http.request.uri.query.tid}) <= 10",
		},
		{
			inPlugin: &olaf.Plugin{
				Type: olaf.PluginTypeCanary,
				Config: map[string]interface{}{
					"upstream":  "staging",
					"key":       "body.tid",
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
			wantMatchExpression:   "int({http.request.body.tid}) > 0 && int({http.request.body.tid}) <= 10",
			wantCanaryFieldInBody: true,
		},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			routes, canaryFieldInBody := canaryReverseProxy(c.inPlugin, c.inServices)
			matchList := routes[0]["match"].([]map[string]string)
			gotMatchExpression := matchList[0]["expression"]

			if gotMatchExpression != c.wantMatchExpression {
				t.Fatalf("Routes: got (%#v), want (%#v)", gotMatchExpression, c.wantMatchExpression)
			}
			if canaryFieldInBody != c.wantCanaryFieldInBody {
				t.Fatalf("CanaryFieldInBody: got (%#v), want (%#v)", canaryFieldInBody, c.wantCanaryFieldInBody)
			}
		})
	}
}
