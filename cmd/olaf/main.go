package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/RussellLuo/olaf/admin"
	"github.com/RussellLuo/olaf/store/yaml"
)

var (
	httpAddr   string
	configFile string
)

func main() {
	flag.StringVar(&httpAddr, "addr", ":2020", "HTTP listen address")
	flag.StringVar(&configFile, "config", "../../caddyconfig/adapter/apis.yaml", "Olaf config file")
	flag.Parse()

	store := yaml.New(configFile)
	server := &http.Server{
		Addr:    httpAddr,
		Handler: admin.NewHTTPRouter(store, admin.NewCodecs()),
	}

	errs := make(chan error, 2)
	go func() {
		log.Printf("transport=HTTP addr=%s\n", httpAddr)
		errs <- server.ListenAndServe()
	}()
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		sig := <-c

		server.Shutdown(context.Background()) // nolint:errcheck
		errs <- fmt.Errorf("%s", sig)
	}()

	log.Printf("terminated, err:%v", <-errs)
}
