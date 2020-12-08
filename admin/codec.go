package admin

import (
	"errors"
	"net/http"

	httpcodec "github.com/RussellLuo/kok/pkg/codec/httpv2"
)

var (
	ErrServiceExists   = errors.New("service already exists")
	ErrServiceNotFound = errors.New("service not found")
	ErrRouteExists     = errors.New("route already exists")
	ErrRouteNotFound   = errors.New("route not found")
	ErrPluginExists    = errors.New("plugin already exists")
	ErrPluginNotFound  = errors.New("plugin not found")

	ErrMethodNotAllowed = errors.New("method not allowed")
	ErrBadRequest       = errors.New("bad request")
)

type Codec struct {
	httpcodec.JSONCodec
}

func (c Codec) EncodeFailureResponse(w http.ResponseWriter, err error) error {
	return c.JSONCodec.EncodeSuccessResponse(w, codeFrom(err), map[string]string{
		"error": err.Error(),
	})
}

func codeFrom(err error) int {
	switch err {
	case ErrServiceExists, ErrRouteExists, ErrPluginExists, ErrBadRequest:
		return http.StatusBadRequest
	case ErrServiceNotFound, ErrRouteNotFound, ErrPluginNotFound:
		return http.StatusNotFound
	case ErrMethodNotAllowed:
		return http.StatusMethodNotAllowed
	default:
		return http.StatusInternalServerError
	}
}

func NewCodecs() httpcodec.CodecMap {
	return httpcodec.CodecMap{
		Default: Codec{},
	}
}
