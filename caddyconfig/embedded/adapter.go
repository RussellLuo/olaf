package adapter

import (
	"encoding/json"
	"io/ioutil"

	"github.com/RussellLuo/olaf/caddyconfig/builder"
	"github.com/RussellLuo/olaf/store/yaml"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/mitchellh/mapstructure"
)

func init() {
	caddyconfig.RegisterAdapter("olaf", Adapter{})
}

type Apps struct {
	HTTP struct {
		Servers map[string]struct {
			Listen []string                 `json:"listen"`
			Routes []map[string]interface{} `json:"routes"`
		} `json:"servers"`
	} `json:"http"`
}

// Adapter adapts Olaf YAML config to Caddy JSON.
type Adapter struct{}

// Adapt converts the Olaf YAML config in body to Caddy JSON.
func (Adapter) Adapt(body []byte, options map[string]interface{}) ([]byte, []caddyconfig.Warning, error) {
	caddyfileAdapter := caddyconfig.GetAdapter("caddyfile")
	caddyfileResult, warn, err := caddyfileAdapter.Adapt(body, options)
	if err != nil {
		return nil, warn, err
	}

	result, err := patch(caddyfileResult)
	if err != nil {
		return nil, nil, err
	}

	return result, nil, nil
}

func patch(caddyfileResult []byte) ([]byte, error) {
	config := make(map[string]interface{})
	if err := json.Unmarshal(caddyfileResult, &config); err != nil {
		return nil, err
	}

	apps := new(Apps)
	if err := mapstructure.Decode(config["apps"], apps); err != nil {
		return nil, err
	}

	for _, server := range apps.HTTP.Servers {
		if err := expandOlaf(server.Routes); err != nil {
			return nil, err
		}
	}

	result, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// expandOlaf replaces the original `olaf` handler with a `subroute` handler,
// whose routes are parsed from the associated YAML file.
func expandOlaf(routes []map[string]interface{}) error {
NextRoute:
	for _, r := range routes {
		handle := r["handle"].([]interface{})
		for _, h := range handle {
			h := h.(map[string]interface{})

			switch h["handler"] {
			case "olaf":
				filename := h["filename"].(string)
				olafRoutes, err := buildOlafRoutes(filename)
				if err != nil {
					return err
				}

				// Replace the `olaf` handler with a `subroute` handler.
				delete(h, "filename")
				h["handler"] = "subroute"
				h["routes"] = olafRoutes

				// We assume that there is only one `olaf` handler in the list.
				continue NextRoute

			case "subroute":
				var subRoutes []map[string]interface{}
				if err := mapstructure.Decode(h["routes"], &subRoutes); err != nil {
					return err
				}
				if err := expandOlaf(subRoutes); err != nil {
					return err
				}

				// We assume that there is only one `olaf` handler in the list.
				continue NextRoute
			}
		}
	}
	return nil
}

func buildOlafRoutes(filename string) ([]map[string]interface{}, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	data, err := yaml.Parse(content)
	if err != nil {
		return nil, err
	}

	return builder.BuildRoutes(data), nil
}
