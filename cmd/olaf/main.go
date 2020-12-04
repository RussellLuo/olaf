package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/RussellLuo/appx"
	"github.com/RussellLuo/kok/pkg/oasv2"
	"github.com/RussellLuo/olaf/admin"
	"github.com/RussellLuo/olaf/admin/file"
	"github.com/RussellLuo/olaf/config/reloader"
)

var (
	httpAddr   string
	configFile string
)

func main() {
	flag.StringVar(&httpAddr, "addr", ":2020", "HTTP listen address")
	flag.StringVar(&configFile, "config", "./olaf.json", "Olaf config file")
	flag.Parse()

	store := file.NewStore(configFile)

	appx.MustRegister(
		appx.New("HTTP-server").InitFunc(func(ctx appx.Context) error {
			server := &http.Server{
				Addr: httpAddr,
				Handler: admin.NewHTTPRouterWithOAS(
					store,
					admin.NewCodecs(),
					&oasv2.ResponseSchema{},
				),
			}
			ctx.Lifecycle.Append(appx.Hook{
				OnStart: func(context.Context) error {
					go server.ListenAndServe() // nolint:errcheck
					log.Printf("transport=HTTP addr=%s\n", httpAddr)
					return nil
				},
				OnStop: func(ctx context.Context) error {
					return server.Shutdown(ctx)
				},
			})
			return nil
		}),
	)

	appx.MustRegister(
		appx.New("Caddy-reloader").InitFunc(func(ctx appx.Context) error {
			c := reloader.NewReloader(store, 5*time.Second)
			ctx.Lifecycle.Append(appx.Hook{
				OnStart: func(context.Context) error {
					go c.Start()
					return nil
				},
				OnStop: func(context.Context) error {
					c.Stop()
					return nil
				},
			})
			return nil
		}),
	)

	if err := appx.Install(context.Background()); err != nil {
		log.Printf("err: %v\n", err)
		return
	}
	defer appx.Uninstall()

	sig, err := appx.Run()
	if err != nil {
		log.Printf("err: %v\n", err)
	}
	log.Printf("terminated, err:%v", sig)
}
