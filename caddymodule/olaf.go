package caddymodule

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	caddy.RegisterModule(Olaf{})
	httpcaddyfile.RegisterHandlerDirective("olaf", parseCaddyfile)
}

// Olaf implements a handler that embeds Olaf's declarative configuration, which
// will be expanded later by a config adapter named `olaf`.
type Olaf struct {
	Filename string `json:"filename"`
}

// CaddyModule returns the Caddy module information.
func (Olaf) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.olaf",
		New: func() caddy.Module { return new(Olaf) },
	}
}

// Validate implements caddy.Validator.
func (o *Olaf) Validate() error {
	if o.Filename == "" {
		return fmt.Errorf("empty filename")
	}
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (o *Olaf) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	return next.ServeHTTP(w, r)
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler. Syntax:
//
//    olaf <filename>
//
func (o *Olaf) UnmarshalCaddyfile(d *caddyfile.Dispenser) (err error) {
	if !d.Next() || !d.NextArg() {
		return d.ArgErr()
	}
	path := d.Val()

	if filepath.IsAbs(path) {
		o.Filename = path
		return nil
	}

	// Make the path relative to the current Caddyfile rather than the
	// current working directory.
	absFile, err := filepath.Abs(d.File())
	if err != nil {
		return fmt.Errorf("failed to get absolute path of file: %s: %v", d.File(), err)
	}
	o.Filename = filepath.Join(filepath.Dir(absFile), path)

	return nil
}

// parseCaddyfile sets up a handler for olaf from Caddyfile tokens.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	o := new(Olaf)
	if err := o.UnmarshalCaddyfile(h.Dispenser); err != nil {
		return nil, err
	}
	return o, nil
}

// Interface guards
var (
	_ caddyhttp.MiddlewareHandler = (*Olaf)(nil)
	_ caddyfile.Unmarshaler       = (*Olaf)(nil)
)
