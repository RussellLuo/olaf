package file

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/RussellLuo/olaf/admin"
	"github.com/RussellLuo/olaf/caddie"
)

type Store struct {
	data     *caddie.Data
	filename string

	mu sync.Mutex
}

func NewStore(filename string) *Store {
	s := &Store{
		data: &caddie.Data{
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

func (s *Store) Load(t time.Time) (*caddie.Data, error) {
	f, err := os.Stat(s.filename)
	if err != nil {
		return nil, err
	}

	if !t.IsZero() && !f.ModTime().After(t) {
		// Not modified, no need to load.
		return nil, caddie.ErrUnmodified
	}

	log.Printf("Loading data from file %s", s.filename)

	content, err := ioutil.ReadFile(s.filename)
	if err != nil {
		return nil, err
	}

	data := new(caddie.Data)
	if err := json.Unmarshal(content, data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Store) save() error {
	content, err := json.Marshal(s.data)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(s.filename, content, 0644)
}

func (s *Store) lockUnlock(err *error) func() {
	s.mu.Lock()
	return func() {
		if err != nil && *err == nil {
			s.save()
		}
		s.mu.Unlock()
	}
}

func (s *Store) CreateService(ctx context.Context, svc *admin.Service) (err error) {
	defer s.lockUnlock(&err)()

	if _, ok := s.data.Services[svc.Name]; ok {
		err = admin.ErrServiceExists
		return
	}

	s.data.Services[svc.Name] = svc
	return
}

func (s *Store) ListServices(ctx context.Context) (services []*admin.Service, err error) {
	defer s.lockUnlock(nil)()

	for _, svc := range s.data.Services {
		services = append(services, svc)
	}
	return
}

func (s *Store) GetService(ctx context.Context, name string) (service *admin.Service, err error) {
	defer s.lockUnlock(nil)()

	svc, ok := s.data.Services[name]
	if !ok {
		err = admin.ErrServiceNotFound
		return
	}
	return svc, nil
}

func (s *Store) UpdateService(ctx context.Context, name string, svc *admin.Service) (err error) {
	defer s.lockUnlock(&err)()

	if _, ok := s.data.Services[svc.Name]; !ok {
		err = admin.ErrServiceNotFound
		return
	}

	s.data.Services[svc.Name] = svc
	return
}

func (s *Store) DeleteService(ctx context.Context, name string) (err error) {
	defer s.lockUnlock(&err)()

	if _, ok := s.data.Services[name]; !ok {
		err = admin.ErrServiceNotFound
		return
	}

	delete(s.data.Services, name)
	return
}

func (s *Store) CreateRoute(ctx context.Context, route *admin.Route) (err error) {
	defer s.lockUnlock(&err)()

	if _, ok := s.data.Routes[route.Name]; ok {
		err = admin.ErrRouteExists
		return
	}

	if len(route.Methods)|len(route.Hosts)|len(route.Paths) == 0 {
		// At least one of methods, hosts or paths, must be provided.
		err = admin.ErrBadRequest
		return
	}

	s.data.Routes[route.Name] = route
	return
}

func (s *Store) ListRoutes(ctx context.Context) (routes []*admin.Route, err error) {
	defer s.lockUnlock(nil)()

	for _, r := range s.data.Routes {
		routes = append(routes, r)
	}
	return
}

func (s *Store) GetRoute(ctx context.Context, name string) (route *admin.Route, err error) {
	defer s.lockUnlock(nil)()

	route, ok := s.data.Routes[name]
	if !ok {
		return nil, admin.ErrRouteNotFound
	}
	return route, nil
}

func (s *Store) UpdateRoute(ctx context.Context, name string, route *admin.Route) (err error) {
	defer s.lockUnlock(&err)()

	if _, ok := s.data.Routes[route.Name]; !ok {
		err = admin.ErrRouteNotFound
		return
	}

	s.data.Routes[route.Name] = route
	return
}

func (s *Store) DeleteRoute(ctx context.Context, name string) (err error) {
	defer s.lockUnlock(&err)()

	if _, ok := s.data.Routes[name]; !ok {
		err = admin.ErrRouteNotFound
		return
	}

	delete(s.data.Routes, name)
	return
}

func (s *Store) CreateTenantCanaryPlugin(ctx context.Context, p *admin.TenantCanaryPlugin) (plugin *admin.TenantCanaryPlugin, err error) {
	defer s.lockUnlock(&err)()

	name := p.Name()
	if name == "" {
		err = admin.ErrBadRequest
		return
	}

	if _, ok := s.data.Plugins[name]; ok {
		err = admin.ErrPluginExists
		return
	}

	s.data.Plugins[name] = p
	return p, nil
}

/*func (s *Store) UpdatePlugin(ctx context.Context, name string) (err error) {
	return nil
}*/

func (s *Store) DeletePlugin(ctx context.Context, name string) (err error) {
	defer s.lockUnlock(&err)()

	if _, ok := s.data.Plugins[name]; !ok {
		err = admin.ErrPluginNotFound
		return
	}

	delete(s.data.Plugins, name)
	return
}
