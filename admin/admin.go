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
	Name string `json:"name"`
	URL  string `json:"url"`

	DialTimeout string `json:"dial_timeout"`
	MaxRequests int    `json:"max_requests"`
}

type Route struct {
	ServiceName string `json:"service_name"`

	Name    string   `json:"name"`
	Methods []string `json:"methods"`
	Hosts   []string `json:"hosts"`
	Paths   []string `json:"paths"`

	StripPrefix string `json:"strip_prefix"`
}

type Plugin struct {
	Service string `json:"service"`
	Route   string `json:"route"`
	Enabled bool   `json:"enabled"`
}

type TenantCanaryPlugin struct {
	Plugin

	Config TenantCanaryConfig `json:"config"`
}

type TenantCanaryConfig struct {
	UpstreamURL string `json:"upstream_url"`

	// query, path, header, body
	TenantIDLocation string        `json:"tenant_id_location"`
	TenantIDName     string        `json:"tenant_id_name"`
	TenantIDList     []int         `json:"tenant_id_list"`
	TenantIDRange    TenantIDRange `json:"tenant_id_range"`
}

type TenantIDRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

func (p *TenantCanaryPlugin) Name() string {
	if p == nil {
		return ""
	}

	if p.Route != "" {
		return PluginTenantCanary + "@" + p.Route
	}
	if p.Service != "" {
		return PluginTenantCanary + "@" + p.Service
	}

	return ""
}
