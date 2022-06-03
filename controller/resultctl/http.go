package resultctl

import (
	"github.com/NubeIO/rubix-automater/automater"
	"github.com/NubeIO/rubix-automater/controller"
	"github.com/NubeIO/rubix-automater/pkg/helpers/apperrors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ResultHTTPHandler is an HTTP controller that exposes result endpoints.
type ResultHTTPHandler struct {
	controller.HTTPHandler
	resultService automater.ResultService
}

// NewResultHTTPHandler creates and returns a new ResultHTTPHandler.
func NewResultHTTPHandler(resultService automater.ResultService) *ResultHTTPHandler {
	return &ResultHTTPHandler{
		resultService: resultService,
	}
}

// Get fetches a job result.
func (hdl *ResultHTTPHandler) Get(c *gin.Context) {
	result, err := hdl.resultService.Get(c.Param("uuid"))
	if err != nil {
		switch err.(type) {
		case *apperrors.NotFoundErr:
			hdl.HandleError(c, http.StatusNotFound, err)
			return
		default:
			hdl.HandleError(c, http.StatusInternalServerError, err)
			return
		}
	}
	c.JSON(http.StatusOK, BuildResponseBodyDTO(result))
}

// Delete deletes a job result.
func (hdl *ResultHTTPHandler) Delete(c *gin.Context) {
	err := hdl.resultService.Delete(c.Param("uuid"))
	if err != nil {
		switch err.(type) {
		case *apperrors.NotFoundErr:
			hdl.HandleError(c, http.StatusNotFound, err)
			return
		default:
			hdl.HandleError(c, http.StatusInternalServerError, err)
			return
		}
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}
