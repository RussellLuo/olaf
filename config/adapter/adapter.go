package adapter

import (
	"encoding/json"

	"github.com/RussellLuo/olaf/config"
	"github.com/RussellLuo/olaf/store/yaml"
	"github.com/caddyserver/caddy/v2/caddyconfig"
)

func init() {
	caddyconfig.RegisterAdapter("olaf", Adapter{})
}

// Adapter adapts Olaf YAML config to Caddy JSON.
type Adapter struct{}

// Adapt converts the Olaf YAML config in body to Caddy JSON.
func (Adapter) Adapt(body []byte, options map[string]interface{}) ([]byte, []caddyconfig.Warning, error) {
	data, err := yaml.Parse(body)
	if err != nil {
		return nil, nil, err
	}
	content := config.BuildCaddyConfig(data)

	result, err := json.Marshal(content)
	if err != nil {
		return nil, nil, err
	}
	return result, nil, nil
}
