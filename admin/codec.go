package admin

import (
	"net/http"

	httpcodec "github.com/RussellLuo/kok/pkg/codec/httpv2"
	"github.com/RussellLuo/olaf"
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
	case olaf.ErrServiceExists, olaf.ErrRouteExists, olaf.ErrPluginExists:
		return http.StatusBadRequest
	case olaf.ErrServiceNotFound, olaf.ErrRouteNotFound, olaf.ErrPluginNotFound:
		return http.StatusNotFound
	case olaf.ErrMethodNotImplemented:
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
