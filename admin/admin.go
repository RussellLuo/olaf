package admin

import (
	"context"
)

//go:generate kokgen ./admin.go Admin

const (
	PluginTenantCanary = "tenant_canary"
)

type Admin interface {
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

	// // @kok(op): PUT /plugins/{name}
	// UpdatePlugin(ctx context.Context, name string) (err error)

	// @kok(op): DELETE /plugins/{name}
	// @kok(success): statusCode:204
	DeletePlugin(ctx context.Context, name string) (err error)
}

type Service struct {
	Name string `json:"name" yaml:"name"`
	URL  string `json:"url" yaml:"url"`

	DialTimeout string `json:"dial_timeout" yaml:"dial_timeout"`
	MaxRequests int    `json:"max_requests" yaml:"max_requests"`
}

type Route struct {
	ServiceName string `json:"service_name" yaml:"service_name"`

	Name    string   `json:"name" yaml:"name"`
	Methods []string `json:"methods" yaml:"methods"`
	Hosts   []string `json:"hosts" yaml:"hosts"`
	Paths   []string `json:"paths" yaml:"paths"`

	StripPrefix string `json:"strip_prefix" yaml:"strip_prefix"`
	AddPrefix   string `json:"add_prefix" yaml:"add_prefix"`
}

type Plugin struct {
	Name    string `json:"name" yaml:"name"`
	Enabled bool   `json:"enabled" yaml:"enabled"`

	Service string `json:"service" yaml:"service"`
	Route   string `json:"route" yaml:"route"`
}

type TenantCanaryPlugin struct {
	Plugin `yaml:",inline"`

	Config TenantCanaryConfig `json:"config" yaml:"config"`
}

type TenantCanaryConfig struct {
	UpstreamURL         string `json:"upstream_url" yaml:"upstream_url"`
	UpstreamDialTimeout string `json:"upstream_dial_timeout" yaml:"upstream_dial_timeout"`
	UpstreamMaxRequests int    `json:"upstream_max_requests" yaml:"upstream_max_requests"`

	// query, path, header, body
	TenantIDLocation  string `json:"tenant_id_location" yaml:"tenant_id_location"`
	TenantIDName      string `json:"tenant_id_name" yaml:"tenant_id_name"`
	TenantIDWhitelist string `json:"tenant_id_whitelist" yaml:"tenant_id_whitelist"`
}

type TenantIDRange struct {
	Start int `json:"start" yaml:"start"`
	End   int `json:"end" yaml:"end"`
}
