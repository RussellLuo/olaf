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

	"github.com/RussellLuo/olaf"
	"github.com/RussellLuo/olaf/caddyconfig/builder"
)

type Loader interface {
	// Load loads the latest data.
	//
	// For efficiency, Implementations should follow the If-Modified-Since style,
	// see https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-Modified-Since
	Load(t time.Time) (*olaf.Data, error)
}

type Reloader struct {
	loader     Loader
	interval   time.Duration
	lastSynced time.Time

	stopC chan struct{}
	exitC chan struct{}
}

func New(loader Loader, interval time.Duration) *Reloader {
	return &Reloader{
		loader:   loader,
		interval: interval,
		stopC:    make(chan struct{}),
		exitC:    make(chan struct{}),
	}
}

func (r *Reloader) Start() {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			data, err := r.loader.Load(r.lastSynced)
			if err != nil {
				if err != olaf.ErrDataUnmodified {
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

func (r *Reloader) reloadCaddy(data *olaf.Data) error {
	content, err := builder.Build(data)
	if err != nil {
		return err
	}

	err = setCaddyConfig(content)
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
