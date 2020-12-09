# Caddy Config Adapter for Olaf

## Build Caddy

Install xcaddy:

```bash
$ go get -u github.com/caddyserver/xcaddy/cmd/xcaddy
```

Build Caddy:

```bash
$ xcaddy build \
    --with github.com/RussellLuo/olaf/config/adapter \
    --with github.com/RussellLuo/caddy-requestbodyvar
```

## Run Caddy

```bash
$ ./caddy run --config olaf.yaml --adapter olaf
```

## Reload Config

```bash
$ ./caddy reload --config olaf.yaml --adapter olaf
```