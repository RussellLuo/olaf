package admin

import (
	"context"

	"github.com/RussellLuo/olaf"
)

//go:generate kokgen ./admin.go Admin

type Admin interface {
	// @kok(op): PUT /server
	// @kok(body): server
	UpdateServer(ctx context.Context, server *olaf.Server) (err error)

	// @kok(op): GET /server
	// @kok(success): body:server
	GetServer(ctx context.Context) (server *olaf.Server, err error)

	// @kok(op): POST /services
	// @kok(body): svc
	CreateService(ctx context.Context, svc *olaf.Service) (err error)

	// @kok(op): GET /services
	// @kok(success): body:services
	ListServices(ctx context.Context) (services []*olaf.Service, err error)

	// @kok(op): GET /services/{name}
	// @kok(success): body:service
	GetService(ctx context.Context, name string) (service *olaf.Service, err error)

	// @kok(op): PUT /services/{name}
	// @kok(body): svc
	UpdateService(ctx context.Context, name string, svc *olaf.Service) (err error)

	// @kok(op): DELETE /services/{name}
	// @kok(success): statusCode:204
	DeleteService(ctx context.Context, name string) (err error)

	// @kok(op): POST /routes
	// @kok(body): route
	CreateRoute(ctx context.Context, route *olaf.Route) (err error)

	// @kok(op): GET /routes
	// @kok(success): body:routes
	ListRoutes(ctx context.Context) (routes []*olaf.Route, err error)

	// @kok(op): GET /routes/{name}
	// @kok(success): body:route
	GetRoute(ctx context.Context, name string) (route *olaf.Route, err error)

	// @kok(op): PUT /routes/{name}
	// @kok(body): route
	UpdateRoute(ctx context.Context, name string, route *olaf.Route) (err error)

	// @kok(op): DELETE /routes/{name}
	// @kok(success): statusCode:204
	DeleteRoute(ctx context.Context, name string) (err error)

	// @kok(op): POST /plugins
	// @kok(body): p
	// @kok(success): body:plugin
	CreatePlugin(ctx context.Context, p *olaf.Plugin) (plugin *olaf.Plugin, err error)

	// @kok(op): GET /plugins
	// @kok(success): body:plugins
	ListPlugins(ctx context.Context) (plugins []*olaf.Plugin, err error)

	// @kok(op): GET /plugins/{name}
	// @kok(success): body:plugin
	GetPlugin(ctx context.Context, name string) (plugin *olaf.Plugin, err error)

	// // @kok(op): PUT /plugins/{name}
	// UpdatePlugin(ctx context.Context, name string) (err error)

	// @kok(op): DELETE /plugins/{name}
	// @kok(success): statusCode:204
	DeletePlugin(ctx context.Context, name string) (err error)
}
