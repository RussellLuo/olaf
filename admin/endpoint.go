// Code generated by kok; DO NOT EDIT.
// github.com/RussellLuo/kok

package admin

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

type CreateRouteRequest struct {
	Route *Route `json:"route"`
}

type CreateRouteResponse struct {
	Err error `json:"-"`
}

func (r *CreateRouteResponse) Body() interface{} { return r }

// Failed implements endpoint.Failer.
func (r *CreateRouteResponse) Failed() error { return r.Err }

// MakeEndpointOfCreateRoute creates the endpoint for s.CreateRoute.
func MakeEndpointOfCreateRoute(s Admin) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*CreateRouteRequest)
		err := s.CreateRoute(
			ctx,
			req.Route,
		)
		return &CreateRouteResponse{
			Err: err,
		}, nil
	}
}

type CreateServiceRequest struct {
	Svc *Service `json:"svc"`
}

type CreateServiceResponse struct {
	Err error `json:"-"`
}

func (r *CreateServiceResponse) Body() interface{} { return r }

// Failed implements endpoint.Failer.
func (r *CreateServiceResponse) Failed() error { return r.Err }

// MakeEndpointOfCreateService creates the endpoint for s.CreateService.
func MakeEndpointOfCreateService(s Admin) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*CreateServiceRequest)
		err := s.CreateService(
			ctx,
			req.Svc,
		)
		return &CreateServiceResponse{
			Err: err,
		}, nil
	}
}

type CreateTenantCanaryPluginRequest struct {
	P *TenantCanaryPlugin `json:"p"`
}

type CreateTenantCanaryPluginResponse struct {
	Plugin *TenantCanaryPlugin `json:"plugin"`
	Err    error               `json:"-"`
}

func (r *CreateTenantCanaryPluginResponse) Body() interface{} { return r.Plugin }

// Failed implements endpoint.Failer.
func (r *CreateTenantCanaryPluginResponse) Failed() error { return r.Err }

// MakeEndpointOfCreateTenantCanaryPlugin creates the endpoint for s.CreateTenantCanaryPlugin.
func MakeEndpointOfCreateTenantCanaryPlugin(s Admin) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*CreateTenantCanaryPluginRequest)
		plugin, err := s.CreateTenantCanaryPlugin(
			ctx,
			req.P,
		)
		return &CreateTenantCanaryPluginResponse{
			Plugin: plugin,
			Err:    err,
		}, nil
	}
}

type DeletePluginRequest struct {
	Name string `json:"-"`
}

type DeletePluginResponse struct {
	Err error `json:"-"`
}

func (r *DeletePluginResponse) Body() interface{} { return r }

// Failed implements endpoint.Failer.
func (r *DeletePluginResponse) Failed() error { return r.Err }

// MakeEndpointOfDeletePlugin creates the endpoint for s.DeletePlugin.
func MakeEndpointOfDeletePlugin(s Admin) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*DeletePluginRequest)
		err := s.DeletePlugin(
			ctx,
			req.Name,
		)
		return &DeletePluginResponse{
			Err: err,
		}, nil
	}
}

type DeleteRouteRequest struct {
	Name string `json:"-"`
}

type DeleteRouteResponse struct {
	Err error `json:"-"`
}

func (r *DeleteRouteResponse) Body() interface{} { return r }

// Failed implements endpoint.Failer.
func (r *DeleteRouteResponse) Failed() error { return r.Err }

// MakeEndpointOfDeleteRoute creates the endpoint for s.DeleteRoute.
func MakeEndpointOfDeleteRoute(s Admin) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*DeleteRouteRequest)
		err := s.DeleteRoute(
			ctx,
			req.Name,
		)
		return &DeleteRouteResponse{
			Err: err,
		}, nil
	}
}

type DeleteServiceRequest struct {
	Name string `json:"-"`
}

type DeleteServiceResponse struct {
	Err error `json:"-"`
}

func (r *DeleteServiceResponse) Body() interface{} { return r }

// Failed implements endpoint.Failer.
func (r *DeleteServiceResponse) Failed() error { return r.Err }

// MakeEndpointOfDeleteService creates the endpoint for s.DeleteService.
func MakeEndpointOfDeleteService(s Admin) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*DeleteServiceRequest)
		err := s.DeleteService(
			ctx,
			req.Name,
		)
		return &DeleteServiceResponse{
			Err: err,
		}, nil
	}
}

type GetRouteRequest struct {
	Name string `json:"-"`
}

type GetRouteResponse struct {
	Route *Route `json:"route"`
	Err   error  `json:"-"`
}

func (r *GetRouteResponse) Body() interface{} { return r.Route }

// Failed implements endpoint.Failer.
func (r *GetRouteResponse) Failed() error { return r.Err }

// MakeEndpointOfGetRoute creates the endpoint for s.GetRoute.
func MakeEndpointOfGetRoute(s Admin) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*GetRouteRequest)
		route, err := s.GetRoute(
			ctx,
			req.Name,
		)
		return &GetRouteResponse{
			Route: route,
			Err:   err,
		}, nil
	}
}

type GetServiceRequest struct {
	Name string `json:"-"`
}

type GetServiceResponse struct {
	Service *Service `json:"service"`
	Err     error    `json:"-"`
}

func (r *GetServiceResponse) Body() interface{} { return r.Service }

// Failed implements endpoint.Failer.
func (r *GetServiceResponse) Failed() error { return r.Err }

// MakeEndpointOfGetService creates the endpoint for s.GetService.
func MakeEndpointOfGetService(s Admin) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*GetServiceRequest)
		service, err := s.GetService(
			ctx,
			req.Name,
		)
		return &GetServiceResponse{
			Service: service,
			Err:     err,
		}, nil
	}
}

type ListRoutesResponse struct {
	Routes []*Route `json:"routes"`
	Err    error    `json:"-"`
}

func (r *ListRoutesResponse) Body() interface{} { return r.Routes }

// Failed implements endpoint.Failer.
func (r *ListRoutesResponse) Failed() error { return r.Err }

// MakeEndpointOfListRoutes creates the endpoint for s.ListRoutes.
func MakeEndpointOfListRoutes(s Admin) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		routes, err := s.ListRoutes(
			ctx,
		)
		return &ListRoutesResponse{
			Routes: routes,
			Err:    err,
		}, nil
	}
}

type ListServicesResponse struct {
	Services []*Service `json:"services"`
	Err      error      `json:"-"`
}

func (r *ListServicesResponse) Body() interface{} { return r.Services }

// Failed implements endpoint.Failer.
func (r *ListServicesResponse) Failed() error { return r.Err }

// MakeEndpointOfListServices creates the endpoint for s.ListServices.
func MakeEndpointOfListServices(s Admin) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		services, err := s.ListServices(
			ctx,
		)
		return &ListServicesResponse{
			Services: services,
			Err:      err,
		}, nil
	}
}

type UpdateRouteRequest struct {
	Name  string `json:"-"`
	Route *Route `json:"route"`
}

type UpdateRouteResponse struct {
	Err error `json:"-"`
}

func (r *UpdateRouteResponse) Body() interface{} { return r }

// Failed implements endpoint.Failer.
func (r *UpdateRouteResponse) Failed() error { return r.Err }

// MakeEndpointOfUpdateRoute creates the endpoint for s.UpdateRoute.
func MakeEndpointOfUpdateRoute(s Admin) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*UpdateRouteRequest)
		err := s.UpdateRoute(
			ctx,
			req.Name,
			req.Route,
		)
		return &UpdateRouteResponse{
			Err: err,
		}, nil
	}
}

type UpdateServiceRequest struct {
	Name string   `json:"-"`
	Svc  *Service `json:"svc"`
}

type UpdateServiceResponse struct {
	Err error `json:"-"`
}

func (r *UpdateServiceResponse) Body() interface{} { return r }

// Failed implements endpoint.Failer.
func (r *UpdateServiceResponse) Failed() error { return r.Err }

// MakeEndpointOfUpdateService creates the endpoint for s.UpdateService.
func MakeEndpointOfUpdateService(s Admin) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*UpdateServiceRequest)
		err := s.UpdateService(
			ctx,
			req.Name,
			req.Svc,
		)
		return &UpdateServiceResponse{
			Err: err,
		}, nil
	}
}
