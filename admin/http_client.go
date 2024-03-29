// Code generated by kun; DO NOT EDIT.
// github.com/RussellLuo/kun

package admin

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/RussellLuo/kun/pkg/httpcodec"
	"github.com/RussellLuo/olaf"
)

type HTTPClient struct {
	codecs     httpcodec.Codecs
	httpClient *http.Client
	scheme     string
	host       string
	pathPrefix string
}

func NewHTTPClient(codecs httpcodec.Codecs, httpClient *http.Client, baseURL string) (*HTTPClient, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	return &HTTPClient{
		codecs:     codecs,
		httpClient: httpClient,
		scheme:     u.Scheme,
		host:       u.Host,
		pathPrefix: strings.TrimSuffix(u.Path, "/"),
	}, nil
}

func (c *HTTPClient) CreatePlugin(ctx context.Context, serviceName string, routeName string, p *olaf.Plugin) (plugin *olaf.Plugin, err error) {
	codec := c.codecs.EncodeDecoder("CreatePlugin")

	path := "/plugins"
	u := &url.URL{
		Scheme: c.scheme,
		Host:   c.host,
		Path:   c.pathPrefix + path,
	}

	reqBody := p
	reqBodyReader, headers, err := codec.EncodeRequestBody(&reqBody)
	if err != nil {
		return nil, err
	}

	_req, err := http.NewRequest("POST", u.String(), reqBodyReader)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		_req.Header.Set(k, v)
	}

	_resp, err := c.httpClient.Do(_req)
	if err != nil {
		return nil, err
	}
	defer _resp.Body.Close()

	if _resp.StatusCode < http.StatusOK || _resp.StatusCode > http.StatusNoContent {
		var respErr error
		err := codec.DecodeFailureResponse(_resp.Body, &respErr)
		if err == nil {
			err = respErr
		}
		return nil, err
	}

	respBody := &CreatePluginResponse{}
	err = codec.DecodeSuccessResponse(_resp.Body, respBody.Body())
	if err != nil {
		return nil, err
	}
	return respBody.Plugin, nil
}

func (c *HTTPClient) CreateRoute(ctx context.Context, serviceName string, route *olaf.Route) (err error) {
	codec := c.codecs.EncodeDecoder("CreateRoute")

	path := "/routes"
	u := &url.URL{
		Scheme: c.scheme,
		Host:   c.host,
		Path:   c.pathPrefix + path,
	}

	reqBody := route
	reqBodyReader, headers, err := codec.EncodeRequestBody(&reqBody)
	if err != nil {
		return err
	}

	_req, err := http.NewRequest("POST", u.String(), reqBodyReader)
	if err != nil {
		return err
	}

	for k, v := range headers {
		_req.Header.Set(k, v)
	}

	_resp, err := c.httpClient.Do(_req)
	if err != nil {
		return err
	}
	defer _resp.Body.Close()

	if _resp.StatusCode < http.StatusOK || _resp.StatusCode > http.StatusNoContent {
		var respErr error
		err := codec.DecodeFailureResponse(_resp.Body, &respErr)
		if err == nil {
			err = respErr
		}
		return err
	}

	return nil
}

func (c *HTTPClient) CreateService(ctx context.Context, svc *olaf.Service) (err error) {
	codec := c.codecs.EncodeDecoder("CreateService")

	path := "/services"
	u := &url.URL{
		Scheme: c.scheme,
		Host:   c.host,
		Path:   c.pathPrefix + path,
	}

	reqBody := svc
	reqBodyReader, headers, err := codec.EncodeRequestBody(&reqBody)
	if err != nil {
		return err
	}

	_req, err := http.NewRequest("POST", u.String(), reqBodyReader)
	if err != nil {
		return err
	}

	for k, v := range headers {
		_req.Header.Set(k, v)
	}

	_resp, err := c.httpClient.Do(_req)
	if err != nil {
		return err
	}
	defer _resp.Body.Close()

	if _resp.StatusCode < http.StatusOK || _resp.StatusCode > http.StatusNoContent {
		var respErr error
		err := codec.DecodeFailureResponse(_resp.Body, &respErr)
		if err == nil {
			err = respErr
		}
		return err
	}

	return nil
}

func (c *HTTPClient) DeletePlugin(ctx context.Context, serviceName string, routeName string, pluginName string) (err error) {
	codec := c.codecs.EncodeDecoder("DeletePlugin")

	path := fmt.Sprintf("/plugins/%s",
		codec.EncodeRequestParam("pluginName", pluginName)[0],
	)
	u := &url.URL{
		Scheme: c.scheme,
		Host:   c.host,
		Path:   c.pathPrefix + path,
	}

	_req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		return err
	}

	_resp, err := c.httpClient.Do(_req)
	if err != nil {
		return err
	}
	defer _resp.Body.Close()

	if _resp.StatusCode < http.StatusOK || _resp.StatusCode > http.StatusNoContent {
		var respErr error
		err := codec.DecodeFailureResponse(_resp.Body, &respErr)
		if err == nil {
			err = respErr
		}
		return err
	}

	return nil
}

func (c *HTTPClient) DeleteRoute(ctx context.Context, serviceName string, routeName string) (err error) {
	codec := c.codecs.EncodeDecoder("DeleteRoute")

	path := fmt.Sprintf("/routes/%s",
		codec.EncodeRequestParam("routeName", routeName)[0],
	)
	u := &url.URL{
		Scheme: c.scheme,
		Host:   c.host,
		Path:   c.pathPrefix + path,
	}

	_req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		return err
	}

	_resp, err := c.httpClient.Do(_req)
	if err != nil {
		return err
	}
	defer _resp.Body.Close()

	if _resp.StatusCode < http.StatusOK || _resp.StatusCode > http.StatusNoContent {
		var respErr error
		err := codec.DecodeFailureResponse(_resp.Body, &respErr)
		if err == nil {
			err = respErr
		}
		return err
	}

	return nil
}

func (c *HTTPClient) DeleteService(ctx context.Context, serviceName string, routeName string) (err error) {
	codec := c.codecs.EncodeDecoder("DeleteService")

	path := fmt.Sprintf("/services/%s",
		codec.EncodeRequestParam("serviceName", serviceName)[0],
	)
	u := &url.URL{
		Scheme: c.scheme,
		Host:   c.host,
		Path:   c.pathPrefix + path,
	}

	_req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		return err
	}

	_resp, err := c.httpClient.Do(_req)
	if err != nil {
		return err
	}
	defer _resp.Body.Close()

	if _resp.StatusCode < http.StatusOK || _resp.StatusCode > http.StatusNoContent {
		var respErr error
		err := codec.DecodeFailureResponse(_resp.Body, &respErr)
		if err == nil {
			err = respErr
		}
		return err
	}

	return nil
}

func (c *HTTPClient) GetConfig(ctx context.Context) (data *olaf.Data, err error) {
	codec := c.codecs.EncodeDecoder("GetConfig")

	path := "/config"
	u := &url.URL{
		Scheme: c.scheme,
		Host:   c.host,
		Path:   c.pathPrefix + path,
	}

	_req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	_resp, err := c.httpClient.Do(_req)
	if err != nil {
		return nil, err
	}
	defer _resp.Body.Close()

	if _resp.StatusCode < http.StatusOK || _resp.StatusCode > http.StatusNoContent {
		var respErr error
		err := codec.DecodeFailureResponse(_resp.Body, &respErr)
		if err == nil {
			err = respErr
		}
		return nil, err
	}

	respBody := &GetConfigResponse{}
	err = codec.DecodeSuccessResponse(_resp.Body, respBody.Body())
	if err != nil {
		return nil, err
	}
	return respBody.Data, nil
}

func (c *HTTPClient) GetPlugin(ctx context.Context, serviceName string, routeName string, pluginName string) (plugin *olaf.Plugin, err error) {
	codec := c.codecs.EncodeDecoder("GetPlugin")

	path := fmt.Sprintf("/plugins/%s",
		codec.EncodeRequestParam("pluginName", pluginName)[0],
	)
	u := &url.URL{
		Scheme: c.scheme,
		Host:   c.host,
		Path:   c.pathPrefix + path,
	}

	_req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	_resp, err := c.httpClient.Do(_req)
	if err != nil {
		return nil, err
	}
	defer _resp.Body.Close()

	if _resp.StatusCode < http.StatusOK || _resp.StatusCode > http.StatusNoContent {
		var respErr error
		err := codec.DecodeFailureResponse(_resp.Body, &respErr)
		if err == nil {
			err = respErr
		}
		return nil, err
	}

	respBody := &GetPluginResponse{}
	err = codec.DecodeSuccessResponse(_resp.Body, respBody.Body())
	if err != nil {
		return nil, err
	}
	return respBody.Plugin, nil
}

func (c *HTTPClient) GetRoute(ctx context.Context, serviceName string, routeName string) (route *olaf.Route, err error) {
	codec := c.codecs.EncodeDecoder("GetRoute")

	path := fmt.Sprintf("/routes/%s",
		codec.EncodeRequestParam("routeName", routeName)[0],
	)
	u := &url.URL{
		Scheme: c.scheme,
		Host:   c.host,
		Path:   c.pathPrefix + path,
	}

	_req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	_resp, err := c.httpClient.Do(_req)
	if err != nil {
		return nil, err
	}
	defer _resp.Body.Close()

	if _resp.StatusCode < http.StatusOK || _resp.StatusCode > http.StatusNoContent {
		var respErr error
		err := codec.DecodeFailureResponse(_resp.Body, &respErr)
		if err == nil {
			err = respErr
		}
		return nil, err
	}

	respBody := &GetRouteResponse{}
	err = codec.DecodeSuccessResponse(_resp.Body, respBody.Body())
	if err != nil {
		return nil, err
	}
	return respBody.Route, nil
}

func (c *HTTPClient) GetService(ctx context.Context, serviceName string, routeName string) (service *olaf.Service, err error) {
	codec := c.codecs.EncodeDecoder("GetService")

	path := fmt.Sprintf("/services/%s",
		codec.EncodeRequestParam("serviceName", serviceName)[0],
	)
	u := &url.URL{
		Scheme: c.scheme,
		Host:   c.host,
		Path:   c.pathPrefix + path,
	}

	_req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	_resp, err := c.httpClient.Do(_req)
	if err != nil {
		return nil, err
	}
	defer _resp.Body.Close()

	if _resp.StatusCode < http.StatusOK || _resp.StatusCode > http.StatusNoContent {
		var respErr error
		err := codec.DecodeFailureResponse(_resp.Body, &respErr)
		if err == nil {
			err = respErr
		}
		return nil, err
	}

	respBody := &GetServiceResponse{}
	err = codec.DecodeSuccessResponse(_resp.Body, respBody.Body())
	if err != nil {
		return nil, err
	}
	return respBody.Service, nil
}

func (c *HTTPClient) GetUpstream(ctx context.Context, upstreamName string, serviceName string) (upstream *olaf.Upstream, err error) {
	codec := c.codecs.EncodeDecoder("GetUpstream")

	path := fmt.Sprintf("/upstreams/%s",
		codec.EncodeRequestParam("upstreamName", upstreamName)[0],
	)
	u := &url.URL{
		Scheme: c.scheme,
		Host:   c.host,
		Path:   c.pathPrefix + path,
	}

	_req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	_resp, err := c.httpClient.Do(_req)
	if err != nil {
		return nil, err
	}
	defer _resp.Body.Close()

	if _resp.StatusCode < http.StatusOK || _resp.StatusCode > http.StatusNoContent {
		var respErr error
		err := codec.DecodeFailureResponse(_resp.Body, &respErr)
		if err == nil {
			err = respErr
		}
		return nil, err
	}

	respBody := &GetUpstreamResponse{}
	err = codec.DecodeSuccessResponse(_resp.Body, respBody.Body())
	if err != nil {
		return nil, err
	}
	return respBody.Upstream, nil
}

func (c *HTTPClient) ListPlugins(ctx context.Context, serviceName string, routeName string) (plugins []*olaf.Plugin, err error) {
	codec := c.codecs.EncodeDecoder("ListPlugins")

	path := "/plugins"
	u := &url.URL{
		Scheme: c.scheme,
		Host:   c.host,
		Path:   c.pathPrefix + path,
	}

	_req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	_resp, err := c.httpClient.Do(_req)
	if err != nil {
		return nil, err
	}
	defer _resp.Body.Close()

	if _resp.StatusCode < http.StatusOK || _resp.StatusCode > http.StatusNoContent {
		var respErr error
		err := codec.DecodeFailureResponse(_resp.Body, &respErr)
		if err == nil {
			err = respErr
		}
		return nil, err
	}

	respBody := &ListPluginsResponse{}
	err = codec.DecodeSuccessResponse(_resp.Body, respBody.Body())
	if err != nil {
		return nil, err
	}
	return respBody.Plugins, nil
}

func (c *HTTPClient) ListRoutes(ctx context.Context, serviceName string) (routes []*olaf.Route, err error) {
	codec := c.codecs.EncodeDecoder("ListRoutes")

	path := "/routes"
	u := &url.URL{
		Scheme: c.scheme,
		Host:   c.host,
		Path:   c.pathPrefix + path,
	}

	_req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	_resp, err := c.httpClient.Do(_req)
	if err != nil {
		return nil, err
	}
	defer _resp.Body.Close()

	if _resp.StatusCode < http.StatusOK || _resp.StatusCode > http.StatusNoContent {
		var respErr error
		err := codec.DecodeFailureResponse(_resp.Body, &respErr)
		if err == nil {
			err = respErr
		}
		return nil, err
	}

	respBody := &ListRoutesResponse{}
	err = codec.DecodeSuccessResponse(_resp.Body, respBody.Body())
	if err != nil {
		return nil, err
	}
	return respBody.Routes, nil
}

func (c *HTTPClient) ListServices(ctx context.Context) (services []*olaf.Service, err error) {
	codec := c.codecs.EncodeDecoder("ListServices")

	path := "/services"
	u := &url.URL{
		Scheme: c.scheme,
		Host:   c.host,
		Path:   c.pathPrefix + path,
	}

	_req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	_resp, err := c.httpClient.Do(_req)
	if err != nil {
		return nil, err
	}
	defer _resp.Body.Close()

	if _resp.StatusCode < http.StatusOK || _resp.StatusCode > http.StatusNoContent {
		var respErr error
		err := codec.DecodeFailureResponse(_resp.Body, &respErr)
		if err == nil {
			err = respErr
		}
		return nil, err
	}

	respBody := &ListServicesResponse{}
	err = codec.DecodeSuccessResponse(_resp.Body, respBody.Body())
	if err != nil {
		return nil, err
	}
	return respBody.Services, nil
}

func (c *HTTPClient) ListUpstreams(ctx context.Context) (upstreams []*olaf.Upstream, err error) {
	codec := c.codecs.EncodeDecoder("ListUpstreams")

	path := "/upstreams"
	u := &url.URL{
		Scheme: c.scheme,
		Host:   c.host,
		Path:   c.pathPrefix + path,
	}

	_req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	_resp, err := c.httpClient.Do(_req)
	if err != nil {
		return nil, err
	}
	defer _resp.Body.Close()

	if _resp.StatusCode < http.StatusOK || _resp.StatusCode > http.StatusNoContent {
		var respErr error
		err := codec.DecodeFailureResponse(_resp.Body, &respErr)
		if err == nil {
			err = respErr
		}
		return nil, err
	}

	respBody := &ListUpstreamsResponse{}
	err = codec.DecodeSuccessResponse(_resp.Body, respBody.Body())
	if err != nil {
		return nil, err
	}
	return respBody.Upstreams, nil
}

func (c *HTTPClient) UpdatePlugin(ctx context.Context, serviceName string, routeName string, pluginName string, plugin *olaf.Plugin) (err error) {
	codec := c.codecs.EncodeDecoder("UpdatePlugin")

	path := fmt.Sprintf("/plugins/%s",
		codec.EncodeRequestParam("pluginName", pluginName)[0],
	)
	u := &url.URL{
		Scheme: c.scheme,
		Host:   c.host,
		Path:   c.pathPrefix + path,
	}

	reqBody := plugin
	reqBodyReader, headers, err := codec.EncodeRequestBody(&reqBody)
	if err != nil {
		return err
	}

	_req, err := http.NewRequest("PUT", u.String(), reqBodyReader)
	if err != nil {
		return err
	}

	for k, v := range headers {
		_req.Header.Set(k, v)
	}

	_resp, err := c.httpClient.Do(_req)
	if err != nil {
		return err
	}
	defer _resp.Body.Close()

	if _resp.StatusCode < http.StatusOK || _resp.StatusCode > http.StatusNoContent {
		var respErr error
		err := codec.DecodeFailureResponse(_resp.Body, &respErr)
		if err == nil {
			err = respErr
		}
		return err
	}

	return nil
}

func (c *HTTPClient) UpdateRoute(ctx context.Context, serviceName string, routeName string, route *olaf.Route) (err error) {
	codec := c.codecs.EncodeDecoder("UpdateRoute")

	path := fmt.Sprintf("/routes/%s",
		codec.EncodeRequestParam("routeName", routeName)[0],
	)
	u := &url.URL{
		Scheme: c.scheme,
		Host:   c.host,
		Path:   c.pathPrefix + path,
	}

	reqBody := route
	reqBodyReader, headers, err := codec.EncodeRequestBody(&reqBody)
	if err != nil {
		return err
	}

	_req, err := http.NewRequest("PUT", u.String(), reqBodyReader)
	if err != nil {
		return err
	}

	for k, v := range headers {
		_req.Header.Set(k, v)
	}

	_resp, err := c.httpClient.Do(_req)
	if err != nil {
		return err
	}
	defer _resp.Body.Close()

	if _resp.StatusCode < http.StatusOK || _resp.StatusCode > http.StatusNoContent {
		var respErr error
		err := codec.DecodeFailureResponse(_resp.Body, &respErr)
		if err == nil {
			err = respErr
		}
		return err
	}

	return nil
}

func (c *HTTPClient) UpdateService(ctx context.Context, serviceName string, routeName string, svc *olaf.Service) (err error) {
	codec := c.codecs.EncodeDecoder("UpdateService")

	path := fmt.Sprintf("/services/%s",
		codec.EncodeRequestParam("serviceName", serviceName)[0],
	)
	u := &url.URL{
		Scheme: c.scheme,
		Host:   c.host,
		Path:   c.pathPrefix + path,
	}

	reqBody := svc
	reqBodyReader, headers, err := codec.EncodeRequestBody(&reqBody)
	if err != nil {
		return err
	}

	_req, err := http.NewRequest("PUT", u.String(), reqBodyReader)
	if err != nil {
		return err
	}

	for k, v := range headers {
		_req.Header.Set(k, v)
	}

	_resp, err := c.httpClient.Do(_req)
	if err != nil {
		return err
	}
	defer _resp.Body.Close()

	if _resp.StatusCode < http.StatusOK || _resp.StatusCode > http.StatusNoContent {
		var respErr error
		err := codec.DecodeFailureResponse(_resp.Body, &respErr)
		if err == nil {
			err = respErr
		}
		return err
	}

	return nil
}

func (c *HTTPClient) UpdateUpstream(ctx context.Context, upstreamName string, serviceName string, upstream *olaf.Upstream) (err error) {
	codec := c.codecs.EncodeDecoder("UpdateUpstream")

	path := fmt.Sprintf("/upstreams/%s",
		codec.EncodeRequestParam("upstreamName", upstreamName)[0],
	)
	u := &url.URL{
		Scheme: c.scheme,
		Host:   c.host,
		Path:   c.pathPrefix + path,
	}

	reqBody := upstream
	reqBodyReader, headers, err := codec.EncodeRequestBody(&reqBody)
	if err != nil {
		return err
	}

	_req, err := http.NewRequest("PUT", u.String(), reqBodyReader)
	if err != nil {
		return err
	}

	for k, v := range headers {
		_req.Header.Set(k, v)
	}

	_resp, err := c.httpClient.Do(_req)
	if err != nil {
		return err
	}
	defer _resp.Body.Close()

	if _resp.StatusCode < http.StatusOK || _resp.StatusCode > http.StatusNoContent {
		var respErr error
		err := codec.DecodeFailureResponse(_resp.Body, &respErr)
		if err == nil {
			err = respErr
		}
		return err
	}

	return nil
}
