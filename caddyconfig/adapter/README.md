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
| `enable_debug` | | Whether to enable debug mode, which sets all log levels to DEBUG (use only for debugging). Default: `false`. |
| `default_log` | | The config of Caddy's DefaultLog. |
| `access_log` | | The config of Caddy's AccessLog. |
| `admin` | | The config of Caddy's admin endpoint, which is used to manage Caddy while it is running. |

The [DefaultLog](https://caddyserver.com/docs/json/logging/#docs) entity:

| Attribute | Required | Description |
| --- | --- | --- |
| `output` | | The config of the LogOutput. |
| `level` | | The minimum entry level to log. Default: `"INFO"`. |

The [AccessLog](https://caddyserver.com/docs/caddyfile/directives/log) entity:

| Attribute | Required | Description |
| --- | --- | --- |
| `disabled` | | Whether to disable this AccessLog. Default: `false`. |
| `output` | | The config of the LogOutput. |
| `level` | | The minimum entry level to log. Default: `"INFO"`. |

The [LogOutput](https://caddyserver.com/docs/caddyfile/directives/log#output-modules) entity:

| Attribute | Required | Description |
| --- | --- | --- |
| `output` | | Where to write the logs. Default: `"stderr"` for DefaultLog, or `"stdout"` for AccessLog. |
| `filename` | | The path to the log file if `output` is `"file"`. Default: `""`. |
| `roll_disabled` | | Whether to disable log rolling. Default: `false`. |
| `roll_size_mb` | | The size (in megabytes) at which to roll the log file. Default: `100`. |
| `roll_keep` | | How many log files to keep before deleting the oldest ones. Default: `10`. |
| `roll_keep_days` | | How long (in days) to keep rolled files. Default: `90`. |

The [Admin](https://caddyserver.com/docs/json/admin/) entity:

| Attribute | Required | Description |
| --- | --- | --- |
| `disabled` | | Whether to disable the admin endpoint completely. Default: `false`. |
| `listen` | | The address to which the admin endpoint's listener should bind itself. Default: `"localhost:2019"`. |
| `enforce_origin` | | See [docs](https://caddyserver.com/docs/json/admin/enforce_origin/). |
| `origins` | | See [docs](https://caddyserver.com/docs/json/admin/origins/). |
| `nonpersistent` | | Whether to keep a copy of the active config on disk. Default: `false` (keep a copy). |

The Service entity:

| Attribute | Required | Description |
| --- | --- | --- |
| `name` | | The name of this Service. Default: `"service_<i>"` (`<i>` is the index of this service in the array). |
| `url`	| √ | The [network address to dial](https://caddyserver.com/docs/json/apps/http/servers/routes/handle/reverse_proxy/upstreams/dial/) to connect to this Service. |
| `dial_timeout` | | The [duration string](https://caddyserver.com/docs/json/apps/http/servers/routes/handle/reverse_proxy/transport/http/dial_timeout/), which indicates how long to wait before timing out trying to connect to this Service. Default: `""` (no timeout). |
| `max_requests` | | The [maximum number of simultaneous requests](https://caddyserver.com/docs/json/apps/http/servers/routes/handle/reverse_proxy/upstreams/max_requests/) to allow to this Service. Default: `0` (no limit). |
| `header_up` | | Sets, adds or removes header fields in a request going upstream to the backend (see [docs](https://caddyserver.com/docs/json/apps/http/servers/routes/handle/reverse_proxy/headers/request/)). Default: `{}` (no header manipulation). |
| `header_down` | | Sets, adds or removes header fields in a response header coming downstream from the backend (see [docs](https://caddyserver.com/docs/json/apps/http/servers/routes/handle/reverse_proxy/headers/response/)). Default: `{}` (no header manipulation). |
| `routes` | √ | A list of Routes associated to this Service. Similar to Kong's [Route Object](https://docs.konghq.com/2.2.x/admin-api/#route-object). |
| `plugins` | | A list of Plugins applied to this Service. Default: `[]`. Similar to Kong's [Plugin Object](https://docs.konghq.com/2.2.x/admin-api/#plugin-object). |

The Route entity:

| Attribute | Required | Description |
| --- | --- | --- |
| `name` | | The name of this Route. Default: `"<service_name>_route_<i>"` (`<i>` is the index of this route in the array). |
| `protocol` | | The request [protocol](https://caddyserver.com/docs/caddyfile/matchers#protocol) that matches this Route. Default: `""` (any protocol). |
| `methods` | | A list of [HTTP methods](https://caddyserver.com/docs/caddyfile/matchers#method) that match this Route. Default: `[]` (any HTTP method). |
| `hosts` | | A list of [hosts](https://caddyserver.com/docs/caddyfile/matchers#host) that match this Route. Default: `[]` (any host). |
| `paths` | √ | A list of [URI paths](https://caddyserver.com/docs/caddyfile/matchers#path) that match this Route. A special prefix `~:` means a [regexp path](https://caddyserver.com/docs/caddyfile/matchers#path-regexp). |
| `headers` | | A list of [headers](https://caddyserver.com/docs/caddyfile/matchers#header) that match this Route. Default: `[]` (any header). |
| `strip_prefix` | | The [prefix](https://caddyserver.com/docs/json/apps/http/servers/routes/handle/rewrite/strip_path_prefix/) that needs to be stripped from the request path. Default: `""` (no stripping). |
| `strip_suffix` | | The [suffix](https://caddyserver.com/docs/json/apps/http/servers/routes/handle/rewrite/strip_path_suffix/) that needs to be stripped from the request path. Default: `""` (no stripping). |
| `target_path` | | The final path when the request is proxied to the target service (using `$` as a placeholder for the request path, which may have been stripped). Default: `""` (leave the request path as is, i.e. `"$"`). |
| `add_prefix` | | The prefix that needs to be added to the final path. Default: `""` (no adding). |
| `priority` | | The priority of this Route. Default: `0`. All the services' routes will be matched from highest priority to lowest. |
| `plugins` | | A list of Plugins applied to this Route. Default: `[]`. Similar to Kong's [Plugin Object](https://docs.konghq.com/2.2.x/admin-api/#plugin-object). |
| `response` | | The static response (see `StaticResponse`) for this Route, which indicates that the request will not be proxied to the target service. Default: `{}` (no static response). |

The [StaticResponse](https://caddyserver.com/docs/json/apps/http/servers/routes/handle/static_response/) entity:

| Attribute | Required | Description |
| --- | --- | --- |
| `status_code` | | The HTTP status code to respond with. Default: `200`. |
| `headers` | | The header fields to set on the response. Default: `{}` (no extra header fields). |
| `body` | | The response body to respond with. Default: `""` (no response body). |
| `close` | | Whether to close the client's connection to the server after writing the response. Default: `false`. |

The Plugin entity:

| Attribute | Required | Description |
| --- | --- | --- |
| `disabled` | | Whether this Plugin is disabled. Default: `false`. |
| `name` | | The name of this Plugin. Default: `"plugin_<i>"` for global plugins, `"<service_name>_plugin_<i>"` for service plugins, or `"<route_name>_plugin_<i>"` for route plugins (`<i>` is the index of this plugin in the array). |
| `type` | √ | The type of this Plugin. Available plugin types: `"canary"` (built-in), or `"request_body_var"` (requires the [caddy-ext/requestbodyvar](https://github.com/RussellLuo/caddy-ext/tree/master/requestbodyvar) extension), or `"rate_limit"` (requires the [caddy-ext/ratelimit](https://github.com/RussellLuo/caddy-ext/tree/master/ratelimit) extension). |
| `order_after` | | The order of this Plugin. Default: `""` (the `type` of the previous Plugin, if any, in the Plugin array). |
| `config` | | The configuration of this Plugin. |

The Config of the Canary Plugin:

| Attribute | Required | Description |
| --- | --- | --- |
| `upstream` | √ | The name of the upstream service for this Plugin. |
| `key` | √ | The variable used to differentiate one client from another. Currently supported variables: `"{path.*}"`, `"{query.*}"`, `"{header.*}"`, `"{cookie.*}"` or `"{body.*}"` (requires the [caddy-ext/requestbodyvar](https://github.com/RussellLuo/caddy-ext/tree/master/requestbodyvar) extension). |
| `type` | | The type of key. Default: `""` (string). |
| `whitelist` | √ | The whitelist defined in a [CEL expression](https://caddyserver.com/docs/caddyfile/matchers#expression) (using `$` as a placeholder for the value of key). If the key value is in the whitelist, the corresponding request will be routed to the service specified by `upstream`. |
| `matcher` | | The advanced matcher, which can consist of various [Caddy matchers](https://caddyserver.com/docs/json/apps/http/servers/routes/match/) or your own ones. **NOTE**: `matcher` and (`key`, `type`, `whitelist`) are mutually exclusive. |
| `strip_prefix` | | The [prefix](https://caddyserver.com/docs/json/apps/http/servers/routes/handle/rewrite/strip_path_prefix/) that needs to be stripped from the request path. Default: `""` (no stripping). |
| `strip_suffix` | | The [suffix](https://caddyserver.com/docs/json/apps/http/servers/routes/handle/rewrite/strip_path_suffix/) that needs to be stripped from the request path. Default: `""` (no stripping). |
| `target_path` | | The final path when the request is proxied to the upstream service (using `$` as a placeholder for the request path, which may have been stripped). Default: `""` (leave the request path as is, i.e. `"$"`). |
| `add_prefix` | | The prefix that needs to be added to the final path. Default: `""` (no adding). |

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
    --with github.com/RussellLuo/caddy-ext/requestbodyvar \
    --with github.com/RussellLuo/caddy-ext/ratelimit
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
