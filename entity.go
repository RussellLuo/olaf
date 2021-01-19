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

type StaticResponse struct {
	Methods    []string            `json:"methods" yaml:"methods"`
	Hosts      []string            `json:"hosts" yaml:"hosts"`
	Paths      []string            `json:"paths" yaml:"paths"`
	StatusCode int                 `json:"status_code" yaml:"status_code"`
	Headers    map[string][]string `json:"headers" yaml:"headers"`
	Body       string              `json:"body" yaml:"body"`
	Close      bool                `json:"close" yaml:"close"`
}

func (ss *StaticResponse) Init() {
	if len(ss.Paths) == 0 {
		ss.Paths = []string{"/*"}
	}
	if ss.StatusCode == 0 {
		ss.StatusCode = 200
	}
}

type LogOutput struct {
	Output       string `json:"output" yaml:"output"`
	Filename     string `json:"filename" yaml:"filename"`
	RollDisabled bool   `json:"roll_disabled" yaml:"roll_disabled"`
	RollSizeMB   int    `json:"roll_size_mb" yaml:"roll_size_mb"`
	RollKeep     int    `json:"roll_keep" yaml:"roll_keep"`
	RollKeepDays int    `json:"roll_keep_days" yaml:"roll_keep_days"`
}

func (o *LogOutput) Init() {
	if o.RollSizeMB == 0 {
		o.RollSizeMB = 100
	}
	if o.RollKeep == 0 {
		o.RollKeep = 10
	}
	if o.RollKeepDays == 0 {
		o.RollKeepDays = 90
	}
}

type AccessLog struct {
	Disabled bool      `json:"disabled" yaml:"disabled"`
	Output   LogOutput `json:"output" yaml:"output"`
	Level    string    `json:"level" yaml:"level"`
}

func (a *AccessLog) Init() {
	a.Output.Init()
	if a.Output.Output == "" {
		a.Output.Output = "stdout"
	}

	if a.Level == "" {
		a.Level = "INFO"
	}
}

type Admin struct {
	Disabled      bool     `json:"disabled" yaml:"disabled"`
	Listen        string   `json:"listen" yaml:"listen"`
	EnforceOrigin bool     `json:"enforce_origin" yaml:"enforce_origin"`
	Origins       []string `json:"origins" yaml:"origins"`
	Nonpersistent bool     `json:"nonpersistent" yaml:"nonpersistent"`
}

func (a *Admin) Init() {
	if a.Listen == "" {
		a.Listen = "localhost:2019"
	}
}

type Server struct {
	Listen          []string          `json:"listen" yaml:"listen"`
	HTTPPort        int               `json:"http_port" yaml:"http_port"`
	HTTPSPort       int               `json:"https_port" yaml:"https_port"`
	EnableAutoHTTPS bool              `json:"enable_auto_https" yaml:"enable_auto_https"`
	EnableDebug     bool              `json:"enable_debug" yaml:"enable_debug"`
	AccessLog       AccessLog         `json:"access_log" yaml:"access_log"`
	Admin           Admin             `json:"admin" yaml:"admin"`
	BeforeResponses []*StaticResponse `json:"before_responses" yaml:"before_responses"`
	AfterResponses  []*StaticResponse `json:"after_responses" yaml:"after_responses"`
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

	s.AccessLog.Init()
	s.Admin.Init()

	for _, ss := range s.BeforeResponses {
		ss.Init()
	}
	for _, ss := range s.AfterResponses {
		ss.Init()
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

	TenantIDLocation  string `json:"tenant_id_location" yaml:"tenant_id_location"`
	TenantIDName      string `json:"tenant_id_name" yaml:"tenant_id_name"`
	TenantIDType      string `json:"tenant_id_type" yaml:"tenant_id_type"`
	TenantIDWhitelist string `json:"tenant_id_whitelist" yaml:"tenant_id_whitelist"`
}

type Data struct {
	Server   *Server                        `json:"server" yaml:"server"`
	Services map[string]*Service            `json:"services" yaml:"services"`
	Routes   map[string]*Route              `json:"routes" yaml:"routes"`
	Plugins  map[string]*TenantCanaryPlugin `json:"plugins" yaml:"plugins"`
}
