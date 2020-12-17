package yaml

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/RussellLuo/olaf"
)

type Store struct {
	filename string

	mu   sync.Mutex
	data *olaf.Data
}

func New(filename string) *Store {
	s := &Store{
		data: &olaf.Data{
			Services: make(map[string]*olaf.Service),
			Routes:   make(map[string]*olaf.Route),
			Plugins:  make(map[string]*olaf.TenantCanaryPlugin),
		},
		filename: filename,
	}

	data, err := s.Load(time.Time{})
	if err != nil {
		log.Printf("Loading err: %v\n", err)
		return s
	}

	s.data = data
	return s
}

func (s *Store) Load(t time.Time) (*olaf.Data, error) {
	f, err := os.Stat(s.filename)
	if err != nil {
		return nil, err
	}

	if !t.IsZero() && !f.ModTime().After(t) {
		// Not modified, no need to load.
		return nil, olaf.ErrDataUnmodified
	}

	log.Printf("Loading data from file %s", s.filename)

	content, err := ioutil.ReadFile(s.filename)
	if err != nil {
		return nil, err
	}

	data, err := Parse(content)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *Store) UpdateServer(ctx context.Context, server *olaf.Server) (err error) {
	return olaf.ErrMethodNotImplemented
}

func (s *Store) GetServer(ctx context.Context) (server *olaf.Server, err error) {
	return s.data.Server, nil
}

func (s *Store) CreateService(ctx context.Context, svc *olaf.Service) (err error) {
	return olaf.ErrMethodNotImplemented
}

func (s *Store) ListServices(ctx context.Context) (services []*olaf.Service, err error) {
	for _, svc := range s.data.Services {
		services = append(services, svc)
	}
	return
}

func (s *Store) GetService(ctx context.Context, name string) (service *olaf.Service, err error) {
	svc, ok := s.data.Services[name]
	if !ok {
		err = olaf.ErrServiceNotFound
		return
	}
	return svc, nil
}

func (s *Store) UpdateService(ctx context.Context, name string, svc *olaf.Service) (err error) {
	return olaf.ErrMethodNotImplemented
}

func (s *Store) DeleteService(ctx context.Context, name string) (err error) {
	return olaf.ErrMethodNotImplemented
}

func (s *Store) CreateRoute(ctx context.Context, route *olaf.Route) (err error) {
	return olaf.ErrMethodNotImplemented
}

func (s *Store) ListRoutes(ctx context.Context) (routes []*olaf.Route, err error) {
	for _, r := range s.data.Routes {
		routes = append(routes, r)
	}
	return
}

func (s *Store) GetRoute(ctx context.Context, name string) (route *olaf.Route, err error) {
	route, ok := s.data.Routes[name]
	if !ok {
		return nil, olaf.ErrRouteNotFound
	}
	return route, nil
}

func (s *Store) UpdateRoute(ctx context.Context, name string, route *olaf.Route) (err error) {
	return olaf.ErrMethodNotImplemented
}

func (s *Store) DeleteRoute(ctx context.Context, name string) (err error) {
	return olaf.ErrMethodNotImplemented
}

func (s *Store) CreateTenantCanaryPlugin(ctx context.Context, p *olaf.TenantCanaryPlugin) (plugin *olaf.TenantCanaryPlugin, err error) {
	return nil, olaf.ErrMethodNotImplemented
}

func (s *Store) ListPlugins(ctx context.Context) (plugins []*olaf.TenantCanaryPlugin, err error) {
	for _, p := range s.data.Plugins {
		plugins = append(plugins, p)
	}
	return
}

func (s *Store) GetPlugin(ctx context.Context, name string) (plugin *olaf.TenantCanaryPlugin, err error) {
	plugin, ok := s.data.Plugins[name]
	if !ok {
		return nil, olaf.ErrPluginNotFound
	}
	return plugin, nil
}

func (s *Store) DeletePlugin(ctx context.Context, name string) (err error) {
	return olaf.ErrMethodNotImplemented
}

// Parse recognizes and parses the YAML content.
func Parse(in []byte) (*olaf.Data, error) {
	c := new(content)
	if err := yaml.Unmarshal(in, c); err != nil {
		return nil, err
	}

	c.Server.Init()
	data := &olaf.Data{
		Server:   c.Server,
		Services: make(map[string]*olaf.Service),
		Routes:   make(map[string]*olaf.Route),
		Plugins:  make(map[string]*olaf.TenantCanaryPlugin),
	}

	for i, s := range c.Services { // global services
		if s.Service.Name == "" {
			s.Service.Name = fmt.Sprintf("service_%d", i)
		}
		data.Services[s.Service.Name] = s.Service

		for j, r := range s.Routes { // routes associated to a service
			if r.Route.Name == "" {
				r.Route.Name = fmt.Sprintf("%s_route_%d", s.Service.Name, j)
			}
			r.Route.ServiceName = s.Service.Name
			data.Routes[r.Route.Name] = r.Route

			for k, p := range r.Plugins { // plugins applied to a route
				if p.Name == "" {
					p.Name = fmt.Sprintf("%s_plugin_%d", r.Route.Name, k)
				}
				p.ServiceName = s.Service.Name
				p.RouteName = r.Route.Name
				data.Plugins[p.Name] = p
			}
		}

		for j, p := range s.Plugins { // plugins applied to a service
			if p.Name == "" {
				p.Name = fmt.Sprintf("%s_plugin_%d", s.Service.Name, j)
			}
			p.ServiceName = s.Service.Name
			data.Plugins[p.Name] = p
		}
	}

	for i, p := range c.Plugins { // global plugins
		if p.Name == "" {
			p.Name = fmt.Sprintf("plugin_%d", i)
		}
		data.Plugins[p.Name] = p
	}

	return data, nil
}

type (
	service struct {
		*olaf.Service `yaml:",inline"`

		Routes  []*route                   `yaml:"routes"`
		Plugins []*olaf.TenantCanaryPlugin `yaml:"plugins"`
	}

	route struct {
		*olaf.Route `yaml:",inline"`

		Plugins []*olaf.TenantCanaryPlugin `yaml:"plugins"`
	}

	content struct {
		Server   *olaf.Server               `yaml:"server"`
		Services []*service                 `yaml:"services"`
		Plugins  []*olaf.TenantCanaryPlugin `yaml:"plugins"`
	}
)
