package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/RussellLuo/appx"
	"github.com/RussellLuo/kok/pkg/oasv2"
	"github.com/RussellLuo/olaf/admin"
	"github.com/RussellLuo/olaf/caddie"
	"github.com/RussellLuo/olaf/store/file"
)

var (
	httpAddr  string
	dataFile  string
	servers   string
	httpPort  int
	httpsPort int
)

func main() {
	flag.StringVar(&httpAddr, "addr", ":2020", "HTTP listen address")
	flag.StringVar(&dataFile, "file", "./olaf.json", "JSON data file")
	flag.StringVar(&servers, "caddy.servers", "localhost:8080", "Server addresses (comma-separated)")
	flag.IntVar(&httpPort, "caddy.http", 80, "HTTP port")
	flag.IntVar(&httpsPort, "caddy.https", 443, "HTTPS port")
	flag.Parse()

	store := file.NewStore(dataFile)

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
		appx.New("Caddie").InitFunc(func(ctx appx.Context) error {
			addrs := strings.Split(servers, ",")
			c := caddie.NewCaddie(store, 5*time.Second, caddie.HTTPConfig{
				Servers:   addrs,
				HTTPPort:  httpPort,
				HTTPSPort: httpsPort,
			})
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
