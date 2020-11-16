package caddie

import (
	"errors"
	"log"
	"time"

	"github.com/RussellLuo/olaf/admin"
)

var (
	// A special error indicates that the data has not been modified
	// since the given time.
	ErrUnmodified = errors.New("data not modified")
)

type Data struct {
	Services map[string]*admin.Service            `json:"services"`
	Routes   map[string]*admin.Route              `json:"routes"`
	Plugins  map[string]*admin.TenantCanaryPlugin `json:"plugins"`
}

type Loader interface {
	// Load loads the latest data.
	//
	// For efficiency, Implementations should follow the If-Modified-Since style,
	// see https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-Modified-Since
	Load(t time.Time) (*Data, error)
}

type Caddie struct {
	loader     Loader
	interval   time.Duration
	lastSynced time.Time

	stopC chan struct{}
	exitC chan struct{}
}

func NewCaddie(loader Loader, interval time.Duration) *Caddie {
	return &Caddie{
		loader:   loader,
		interval: interval,
		stopC:    make(chan struct{}),
		exitC:    make(chan struct{}),
	}
}

func (c *Caddie) Start() {
	tickC := time.Tick(c.interval)
	for {
		select {
		case <-tickC:
			data, err := c.loader.Load(c.lastSynced)
			if err != nil {
				if err != ErrUnmodified {
					log.Printf("Load data err: %v\n", err)
				}
				continue
			}

			if err := c.reloadCaddy(data); err != nil {
				log.Printf("Reload Caddy err: %v\n", err)
				continue
			}

			c.lastSynced = time.Now()

		case <-c.stopC:
			goto exit
		}
	}

exit:
	close(c.exitC)
}

func (c *Caddie) Stop() {
	close(c.stopC)
	<-c.exitC
}
