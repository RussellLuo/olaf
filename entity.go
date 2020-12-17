package olaf

import (
	"errors"
)

var (
	ErrServiceExists   = errors.New("service already exists")
	ErrServiceNotFound = errors.New("service not found")

	ErrRouteExists   = errors.New("route already exists")
	ErrRouteNotFound = errors.New("route not found")

	ErrPluginExists   = errors.New("plugin already exists")
	ErrPluginNotFound = errors.New("plugin not found")

	ErrMethodNotImplemented = errors.New("method not implemented")

	// A special error indicates that the data has not been modified
	// since the given time.
	ErrDataUnmodified = errors.New("data unmodified")
)

type Server struct {
	Listen           []string `json:"listen" yaml:"listen"`
	HTTPPort         int      `json:"http_port" yaml:"http_port"`
	HTTPSPort        int      `json:"https_port" yaml:"https_port"`
	EnableAutoHTTPS  bool     `json:"enable_auto_https" yaml:"enable_auto_https"`
	DisableAccessLog bool     `json:"disable_access_log" yaml:"disable_access_log"`
}

func (s *Server) Init() {
	if s == nil {
		return
	}

	if len(s.Listen) == 0 {
		s.Listen = []string{":6060"}
	}
	if s.HTTPPort == 0 {
		s.HTTPPort = 80
	}
	if s.HTTPSPort == 0 {
		s.HTTPSPort = 443
	}
}

type Service struct {
	Name string `json:"name" yaml:"name"`
	URL  string `json:"url" yaml:"url"`

	DialTimeout string `json:"dial_timeout" yaml:"dial_timeout"`
	MaxRequests int    `json:"max_requests" yaml:"max_requests"`
}

type Route struct {
	ServiceName string `json:"service_name" yaml:"service_name"`

	// Route name must be unique.
	Name    string   `json:"name" yaml:"name"`
	Methods []string `json:"methods" yaml:"methods"`
	Hosts   []string `json:"hosts" yaml:"hosts"`
	Paths   []string `json:"paths" yaml:"paths"`

	StripPrefix string `json:"strip_prefix" yaml:"strip_prefix"`
	AddPrefix   string `json:"add_prefix" yaml:"add_prefix"`

	// Routes will be matched from highest priority to lowest.
	Priority int `json:"priority" yaml:"priority"`
}

type Plugin struct {
	Name    string `json:"name" yaml:"name"`
	Enabled bool   `json:"enabled" yaml:"enabled"`

	RouteName   string `json:"route_name" yaml:"route_name"`
	ServiceName string `json:"service_name" yaml:"service_name"`
}

type TenantCanaryPlugin struct {
	Plugin `yaml:",inline"`

	Config TenantCanaryConfig `json:"config" yaml:"config"`
}

type TenantCanaryConfig struct {
	UpstreamServiceName string `json:"upstream_service_name" yaml:"upstream_service_name"`

	// query, path, header, body
	TenantIDLocation  string `json:"tenant_id_location" yaml:"tenant_id_location"`
	TenantIDName      string `json:"tenant_id_name" yaml:"tenant_id_name"`
	TenantIDWhitelist string `json:"tenant_id_whitelist" yaml:"tenant_id_whitelist"`
}

type Data struct {
	Server   *Server                        `json:"server" yaml:"server"`
	Services map[string]*Service            `json:"services" yaml:"services"`
	Routes   map[string]*Route              `json:"routes" yaml:"routes"`
	Plugins  map[string]*TenantCanaryPlugin `json:"plugins" yaml:"plugins"`
}
