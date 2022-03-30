package adapter_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	adapter "github.com/RussellLuo/olaf/caddyconfig/embedded"
)

func TestExpander_Expand(t *testing.T) {
	expander := adapter.NewExpander(nil)

	getConfigFromJSON := func(t *testing.T, jsonPath string) map[string]interface{} {
		content, err := ioutil.ReadFile(jsonPath)
		if err != nil {
			t.Fatalf("err: %v\n", err)
		}

		var config map[string]interface{}
		if err := json.Unmarshal(content, &config); err != nil {
			t.Fatalf("err: %v\n", err)
		}

		return config
	}

	cases := []struct {
		name           string
		inConfigJSON   string
		wantConfigJSON string
		wantErrStr     string
	}{
		{
			name:           "ok",
			inConfigJSON:   "./testdata/in.json",
			wantConfigJSON: "./testdata/want.json",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			inConfig := getConfigFromJSON(t, c.inConfigJSON)
			err := expander.Expand(inConfig)
			if (err == nil && c.wantErrStr != "") || (err != nil && err.Error() != c.wantErrStr) {
				t.Fatalf("Err: got (%#v), want (%#v)", err, c.wantErrStr)
			}

			wantConfig := getConfigFromJSON(t, c.wantConfigJSON)
			if fmt.Sprintf("%+v", inConfig) != fmt.Sprintf("%+v", wantConfig) {
				t.Fatalf("Config: got (%+v), want (%+v)", inConfig, wantConfig)
			}
		})
	}
}
