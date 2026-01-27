package server

import (
	"net/http"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/api/generated/models"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/handler"
	"github.com/alishashelby/Samok-Aah-t/backend/pkg/response"
)

func NewResponseErrorHandler(mapper *handler.ErrorMapper) func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		status, code, msg := mapper.MapError(err)

		resp := models.ErrorResponse{}
		resp.Code = code
		resp.Message = msg

		response.SendJSON(w, status, resp)
	}
}
