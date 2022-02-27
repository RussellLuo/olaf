package admin

import (
	"net/http"

	"github.com/RussellLuo/kun/pkg/httpcodec"
	"github.com/RussellLuo/olaf"
)

type Codec struct {
	httpcodec.JSON
}

func (c Codec) EncodeFailureResponse(w http.ResponseWriter, err error) error {
	return c.JSON.EncodeSuccessResponse(w, codeFrom(err), map[string]string{
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

func NewCodecs() *httpcodec.DefaultCodecs {
	return httpcodec.NewDefaultCodecs(Codec{})
}
