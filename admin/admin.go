package admin

import (
	"context"

	"github.com/RussellLuo/olaf"
)

//go:generate kungen ./admin.go Admin

type Admin interface {
	//kun:op POST /services
	//kun:body svc
	CreateService(ctx context.Context, svc *olaf.Service) (err error)

	//kun:op GET /services
	//kun:success body=services
	ListServices(ctx context.Context) (services []*olaf.Service, err error)

	//kun:op GET /services/{serviceName}
	//kun:op GET /routes/{routeName}/service
	//kun:success body=service
	GetService(ctx context.Context, serviceName, routeName string) (service *olaf.Service, err error)

	//kun:op PUT /services/{serviceName}
	//kun:op PUT /routes/{routeName}/service
	//kun:body svc
	UpdateService(ctx context.Context, serviceName, routeName string, svc *olaf.Service) (err error)

	//kun:op DELETE /services/{serviceName}
	//kun:op DELETE /routes/{routeName}/service
	//kun:success statusCode=204
	DeleteService(ctx context.Context, serviceName, routeName string) (err error)

	//kun:op POST /routes
	//kun:op POST /services/{serviceName}/routes
	//kun:body route
	CreateRoute(ctx context.Context, serviceName string, route *olaf.Route) (err error)

	//kun:op GET /routes
	//kun:op GET /services/{serviceName}/routes
	//kun:success body=routes
	ListRoutes(ctx context.Context, serviceName string) (routes []*olaf.Route, err error)

	//kun:op GET /routes/{routeName}
	//kun:op GET /services/{serviceName}/routes/{routeName}
	//kun:success body=route
	GetRoute(ctx context.Context, serviceName, routeName string) (route *olaf.Route, err error)

	//kun:op PUT /routes/{routeName}
	//kun:op PUT /services/{serviceName}/routes/{routeName}
	//kun:body route
	UpdateRoute(ctx context.Context, serviceName, routeName string, route *olaf.Route) (err error)

	//kun:op DELETE /routes/{routeName}
	//kun:op DELETE /services/{serviceName}/routes/{routeName}
	//kun:success statusCode=204
	DeleteRoute(ctx context.Context, serviceName, routeName string) (err error)

	//kun:op POST /plugins
	//kun:op POST /routes/{routeName}/plugins
	//kun:op POST /services/{serviceName}/plugins
	//kun:body p
	//kun:success body=plugin
	CreatePlugin(ctx context.Context, serviceName, routeName string, p *olaf.Plugin) (plugin *olaf.Plugin, err error)

	//kun:op GET /plugins
	//kun:op GET /routes/{routeName}/plugins
	//kun:op GET /services/{serviceName}/plugins
	//kun:success body=plugins
	ListPlugins(ctx context.Context, serviceName, routeName string) (plugins []*olaf.Plugin, err error)

	//kun:op GET /plugins/{pluginName}
	//kun:op GET /routes/{routeName}/plugins/{pluginName}
	//kun:op GET /services/{serviceName}/plugins/{pluginName}
	//kun:success body=plugin
	GetPlugin(ctx context.Context, serviceName, routeName, pluginName string) (plugin *olaf.Plugin, err error)

	//kun:op PUT /plugins/{pluginName}
	//kun:op PUT /routes/{routeName}/plugins/{pluginName}
	//kun:op PUT /services/{serviceName}/plugins/{pluginName}
	//kun:body plugin
	UpdatePlugin(ctx context.Context, serviceName, routeName, pluginName string, plugin *olaf.Plugin) (err error)

	//kun:op DELETE /plugins/{pluginName}
	//kun:op DELETE /routes/{routeName}/plugins/{pluginName}
	//kun:op DELETE /services/{serviceName}/plugins/{pluginName}
	//kun:success statusCode=204
	DeletePlugin(ctx context.Context, serviceName, routeName, pluginName string) (err error)

	//kun:op POST /upstreams
	//kun:body upstream
	//CreateUpstream(ctx context.Context, upstream *olaf.Upstream) (err error)

	//kun:op GET /upstreams
	//kun:success body=upstreams
	ListUpstreams(ctx context.Context) (upstreams []*olaf.Upstream, err error)

	//kun:op GET /upstreams/{upstreamName}
	//kun:op GET /services/{serviceName}/upstream
	//kun:success body=upstream
	GetUpstream(ctx context.Context, upstreamName, serviceName string) (upstream *olaf.Upstream, err error)

	//kun:op PUT /upstreams/{upstreamName}
	//kun:op PUT /services/{serviceName}/upstream
	//kun:body upstream
	UpdateUpstream(ctx context.Context, upstreamName, serviceName string, upstream *olaf.Upstream) (err error)

	//kun:op DELETE /upstreams/{upstreamName}
	//kun:op DELETE /services/{serviceName}/upstream
	//kun:success statusCode=204
	//DeleteUpstream(ctx context.Context, upstreamName, serviceName string) (err error)
}
