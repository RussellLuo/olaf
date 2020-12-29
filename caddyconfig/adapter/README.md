# Olaf's Declarative Configuration

## Declarative Configuration

### Overview

Olaf's declarative configuration is inspired by Kong, and the configuration must be written in YAML.

For the core idea of the declarative configuration, see [Kong's Declarative Configuration](https://docs.konghq.com/2.2.x/db-less-and-declarative-config/#what-is-declarative-configuration).

### Entities

While following the same idea as Kong, Olaf's declarative configuration and the entities it contains are different in details.

The top-level entries:

| Entry | Required | Description |
| --- | --- | --- |
| `server` | | The Server options. |
| `services` | √ | A list of Services. Similar to Kong's [Service Object](https://docs.konghq.com/2.2.x/admin-api/#service-object). |
| `plugins` | | A list of global Plugins. Default: `[]`. Similar to Kong's [Plugin Object](https://docs.konghq.com/2.2.x/admin-api/#plugin-object). |

The Server entity:

| Attribute | Required | Description |
| --- | --- | --- |
| `listen` | | [Network addresses](https://caddyserver.com/docs/json/apps/http/servers/listen/) to which to bind listeners. Default: `[":6060"]`. |
| `http_port` | | The port to use for [HTTP](https://caddyserver.com/docs/json/apps/http/http_port/). Default: `80`. |
| `https_port` | | The port to use for [HTTPS](https://caddyserver.com/docs/json/apps/http/https_port/). Default: `443`. |
| `enable_auto_https` | | Whether to enable [automatic HTTPS](https://caddyserver.com/docs/json/apps/http/servers/automatic_https/), Default: `false`. |
| `disable_access_log` | | Whether to disable [access logs](https://caddyserver.com/docs/caddyfile/directives/log). Default: `false`. |
| `enable_debug` | | Whether to enable debug mode, which sets all log levels to DEBUG (use only for debugging). Default: `false`. |

The Service entity:

| Attribute | Required | Description |
| --- | --- | --- |
| `name` | | The name of this Service. Default: `"service_<i>"` (`<i>` is the index of this service in the array). |
| `url`	| √ | The [network address to dial](https://caddyserver.com/docs/json/apps/http/servers/routes/handle/reverse_proxy/upstreams/dial/) to connect to this Service. |
| `dial_timeout` | | The [duration string](https://caddyserver.com/docs/json/apps/http/servers/routes/handle/reverse_proxy/transport/http/dial_timeout/), which indicates how long to wait before timing out trying to connect to this Service. Default: `""` (no timeout). |
| `max_requests` | | The [maximum number of simultaneous requests](https://caddyserver.com/docs/json/apps/http/servers/routes/handle/reverse_proxy/upstreams/max_requests/) to allow to this Service. Default: `0` (no limit). |
| `routes` | √ | A list of Routes associated to this Service. Similar to Kong's [Route Object](https://docs.konghq.com/2.2.x/admin-api/#route-object). |
| `plugins` | | A list of Plugins applied to this Service. Default: `[]`. Similar to Kong's [Plugin Object](https://docs.konghq.com/2.2.x/admin-api/#plugin-object). |

The Route entity:

| Attribute | Required | Description |
| --- | --- | --- |
| `name` | | The name of this Route. Default: `"<service_name>_route_<i>"` (`<i>` is the index of this route in the array). |
| `methods` | | A list of [HTTP methods](https://caddyserver.com/docs/caddyfile/matchers#method) that match this Route. Default: `[]` (any HTTP method). |
| `hosts` | | A list of [hosts](https://caddyserver.com/docs/caddyfile/matchers#host) that match this Route. Default: `[]` (any host). |
| `paths` | √ | A list of [URI paths](https://caddyserver.com/docs/caddyfile/matchers#path) that match this Route. A special prefix `~:` means a [regexp path](https://caddyserver.com/docs/caddyfile/matchers#path-regexp). |
| `strip_prefix` | | The prefix that needs to be stripped from the original path. Default: `""` (no stripping). |
| `add_prefix` | | The prefix that needs to be added to the final path. Default: `""` (no adding). |
| `priority` | | The priority of this Route. Default: `0`. Routes will be matched from highest priority to lowest. |
| `plugins` | | A list of plugins applied to this Route. Default: `[]`. Similar to Kong's [Plugin Object](https://docs.konghq.com/2.2.x/admin-api/#plugin-object). |

The Plugin entity:

| Attribute | Required | Description |
| --- | --- | --- |
| `name` | | The name of this Plugin. Default: `"plugin_<i>"` for global plugins, `"<service_name>_plugin_<i>"` for service plugins, or `"<route_name>_plugin_<i>"` for route plugins (`<i>` is the index of this plugin in the array). |
| `enabled` | | Whether this Plugin is applied. Default: `false` |
| `config` | √ | The configuration of this Plugin. |

The Config of the Tenant Canary Plugin:

| Attribute | Required | Description |
| --- | --- | --- |
| `upstream_service_name` | √ | The name of the upstream service for this Plugin. |
| `tenant_id_location` | √ | The location of Tenant-ID in the HTTP request. Options: `"path"`, `"query"`, `"header"`, `"cookie"` or `"body"` (requires the [caddy-requestbodyvar](https://github.com/RussellLuo/caddy-requestbodyvar) extension). |
| `tenant_id_name` | √ | The index of the Tenant-ID part in the path (see [{http.request.uri.path.*}](https://caddyserver.com/docs/json/apps/http/#docs)), if `tenant_id_location` is `"path"`. Otherwise, the name of Tenant-ID in the request. |
| `tenant_id_type` | | The type of Tenant-ID. Default: `""` (string). |
| `tenant_id_whitelist` | √ | The Tenant-ID whitelist defined in a [CEL expression](https://caddyserver.com/docs/caddyfile/matchers#expression) (using `$` as a placeholder for the value of Tenant-ID). If the value of Tenant-ID is in the whitelist, the corresponding request will be routed to the service specified by `upstream_service_name`. |

### Example

See [olaf.yaml](olaf.yaml).


## Usage

### Build Caddy

Install xcaddy:

```bash
$ go get -u github.com/caddyserver/xcaddy/cmd/xcaddy
```

Build Caddy:

```bash
$ xcaddy build \
    --with github.com/RussellLuo/olaf/caddyconfig/adapter \
    --with github.com/RussellLuo/caddy-requestbodyvar
```

### Run Caddy

```bash
$ ./caddy run --config olaf.yaml --adapter olaf
```

### Reload Config

Don't run, just test the configuration:

```bash
$ ./caddy adapt --config olaf.yaml --adapter olaf --validate > /dev/null
```

Reload the configuration:

```bash
$ ./caddy reload --config olaf.yaml --adapter olaf
```
