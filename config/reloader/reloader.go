package reloader

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/RussellLuo/olaf/config"
)

type Loader interface {
	// Load loads the latest data.
	//
	// For efficiency, Implementations should follow the If-Modified-Since style,
	// see https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-Modified-Since
	Load(t time.Time) (*config.Data, error)
}

type Reloader struct {
	loader     Loader
	interval   time.Duration
	lastSynced time.Time

	stopC chan struct{}
	exitC chan struct{}
}

func NewReloader(loader Loader, interval time.Duration) *Reloader {
	return &Reloader{
		loader:   loader,
		interval: interval,
		stopC:    make(chan struct{}),
		exitC:    make(chan struct{}),
	}
}

func (r *Reloader) Start() {
	tickC := time.Tick(r.interval)
	for {
		select {
		case <-tickC:
			data, err := r.loader.Load(r.lastSynced)
			if err != nil {
				if err != config.ErrUnmodified {
					log.Printf("Load data err: %v\n", err)
				}
				continue
			}

			if err := r.reloadCaddy(data); err != nil {
				log.Printf("Reload Caddy err: %v\n", err)
				continue
			}

			r.lastSynced = time.Now()

		case <-r.stopC:
			goto exit
		}
	}

exit:
	close(r.exitC)
}

func (r *Reloader) Stop() {
	close(r.stopC)
	<-r.exitC
}

func (r *Reloader) reloadCaddy(data *config.Data) error {
	content := config.BuildCaddyConfig(data)

	err := setCaddyConfig(content)
	if err != nil {
		log.Printf("new config: %#v\n", content)
	}
	return err
}

func setCaddyConfig(config map[string]interface{}) error {
	u := &url.URL{
		Scheme: "http",
		Host:   "localhost:2019",
		Path:   "/load",
	}

	reqBody, err := json.Marshal(config)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusNoContent {
		msg, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(msg))
	}

	return nil
}
