package admin

import (
	"context"
)

//go:generate kokgen ./admin.go Admin

type Admin interface {
	// @kok(op): PUT /server
	// @kok(body): server
	UpdateServer(ctx context.Context, server *Server) (err error)

	// @kok(op): GET /server
	// @kok(success): body:server
	GetServer(ctx context.Context) (server *Server, err error)

	// @kok(op): POST /services
	// @kok(body): svc
	CreateService(ctx context.Context, svc *Service) (err error)

	// @kok(op): GET /services
	// @kok(success): body:services
	ListServices(ctx context.Context) (services []*Service, err error)

	// @kok(op): GET /services/{name}
	// @kok(success): body:service
	GetService(ctx context.Context, name string) (service *Service, err error)

	// @kok(op): PUT /services/{name}
	// @kok(body): svc
	UpdateService(ctx context.Context, name string, svc *Service) (err error)

	// @kok(op): DELETE /services/{name}
	// @kok(success): statusCode:204
	DeleteService(ctx context.Context, name string) (err error)

	// @kok(op): POST /routes
	// @kok(body): route
	CreateRoute(ctx context.Context, route *Route) (err error)

	// @kok(op): GET /routes
	// @kok(success): body:routes
	ListRoutes(ctx context.Context) (routes []*Route, err error)

	// @kok(op): GET /routes/{name}
	// @kok(success): body:route
	GetRoute(ctx context.Context, name string) (route *Route, err error)

	// @kok(op): PUT /routes/{name}
	// @kok(body): route
	UpdateRoute(ctx context.Context, name string, route *Route) (err error)

	// @kok(op): DELETE /routes/{name}
	// @kok(success): statusCode:204
	DeleteRoute(ctx context.Context, name string) (err error)

	// @kok(op): POST /plugins
	// @kok(body): p
	// @kok(success): body:plugin
	CreateTenantCanaryPlugin(ctx context.Context, p *TenantCanaryPlugin) (plugin *TenantCanaryPlugin, err error)

	// @kok(op): GET /plugins
	// @kok(success): body:plugins
	ListPlugins(ctx context.Context) (plugins []*TenantCanaryPlugin, err error)

	// @kok(op): GET /plugins/{name}
	// @kok(success): body:plugin
	GetPlugin(ctx context.Context, name string) (plugin *TenantCanaryPlugin, err error)

	// // @kok(op): PUT /plugins/{name}
	// UpdatePlugin(ctx context.Context, name string) (err error)

	// @kok(op): DELETE /plugins/{name}
	// @kok(success): statusCode:204
	DeletePlugin(ctx context.Context, name string) (err error)
}

type Server struct {
	Listen           []string `json:"listen" yaml:"listen"`
	HTTPPort         int      `json:"http_port" yaml:"http_port"`
	HTTPSPort        int      `json:"https_port" yaml:"https_port"`
	EnableAutoHTTPS  bool     `json:"enable_auto_https" yaml:"enable_auto_https"`
	DisableAccessLog bool     `json:"disable_access_log" yaml:"disable_access_log"`
}

func (s *Server) Init() {
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
