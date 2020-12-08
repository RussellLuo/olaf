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

	"github.com/RussellLuo/olaf/admin"
	"github.com/RussellLuo/olaf/config"
)

type Store struct {
	filename string

	mu   sync.Mutex
	data *config.Data
}

func New(filename string) *Store {
	s := &Store{
		data: &config.Data{
			Services: make(map[string]*admin.Service),
			Routes:   make(map[string]*admin.Route),
			Plugins:  make(map[string]*admin.TenantCanaryPlugin),
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

func (s *Store) Load(t time.Time) (*config.Data, error) {
	f, err := os.Stat(s.filename)
	if err != nil {
		return nil, err
	}

	if !t.IsZero() && !f.ModTime().After(t) {
		// Not modified, no need to load.
		return nil, config.ErrUnmodified
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

func (s *Store) CreateService(ctx context.Context, svc *admin.Service) (err error) {
	return admin.ErrMethodNotAllowed
}

func (s *Store) ListServices(ctx context.Context) (services []*admin.Service, err error) {
	for _, svc := range s.data.Services {
		services = append(services, svc)
	}
	return
}

func (s *Store) GetService(ctx context.Context, name string) (service *admin.Service, err error) {
	svc, ok := s.data.Services[name]
	if !ok {
		err = admin.ErrServiceNotFound
		return
	}
	return svc, nil
}

func (s *Store) UpdateService(ctx context.Context, name string, svc *admin.Service) (err error) {
	return admin.ErrMethodNotAllowed
}

func (s *Store) DeleteService(ctx context.Context, name string) (err error) {
	return admin.ErrMethodNotAllowed
}

func (s *Store) CreateRoute(ctx context.Context, route *admin.Route) (err error) {
	return admin.ErrMethodNotAllowed
}

func (s *Store) ListRoutes(ctx context.Context) (routes []*admin.Route, err error) {
	for _, r := range s.data.Routes {
		routes = append(routes, r)
	}
	return
}

func (s *Store) GetRoute(ctx context.Context, name string) (route *admin.Route, err error) {
	route, ok := s.data.Routes[name]
	if !ok {
		return nil, admin.ErrRouteNotFound
	}
	return route, nil
}

func (s *Store) UpdateRoute(ctx context.Context, name string, route *admin.Route) (err error) {
	return admin.ErrMethodNotAllowed
}

func (s *Store) DeleteRoute(ctx context.Context, name string) (err error) {
	return admin.ErrMethodNotAllowed
}

func (s *Store) CreateTenantCanaryPlugin(ctx context.Context, p *admin.TenantCanaryPlugin) (plugin *admin.TenantCanaryPlugin, err error) {
	return nil, admin.ErrMethodNotAllowed
}

func (s *Store) DeletePlugin(ctx context.Context, name string) (err error) {
	return admin.ErrMethodNotAllowed
}

// Parse recognizes and parses the YAML content.
func Parse(in []byte) (*config.Data, error) {
	c := new(content)
	if err := yaml.Unmarshal(in, c); err != nil {
		return nil, err
	}

	data := &config.Data{
		Server:   c.Server,
		Services: make(map[string]*admin.Service),
		Routes:   make(map[string]*admin.Route),
		Plugins:  make(map[string]*admin.TenantCanaryPlugin),
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
		*admin.Service `yaml:",inline"`

		Routes  []*route                    `yaml:"routes"`
		Plugins []*admin.TenantCanaryPlugin `yaml:"plugins"`
	}

	route struct {
		*admin.Route `yaml:",inline"`

		Plugins []*admin.TenantCanaryPlugin `yaml:"plugins"`
	}

	content struct {
		Server   config.Server               `yaml:"server"`
		Services []*service                  `yaml:"services"`
		Plugins  []*admin.TenantCanaryPlugin `yaml:"plugins"`
	}
)
