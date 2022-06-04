package admin

import (
	"github.com/NubeIO/rubix-automater/automater"
	"github.com/NubeIO/rubix-automater/controller"
	"github.com/gin-gonic/gin"
	"net/http"
)

// AdminHTTPHandler is an HTTP controller that exposes result endpoints.
type AdminHTTPHandler struct {
	controller.HTTPHandler
	storage automater.Storage
}

// NewAdminHTTPHandler creates and returns a new ResultHTTPHandler.
func NewAdminHTTPHandler(storage automater.Storage) *AdminHTTPHandler {
	return &AdminHTTPHandler{
		storage: storage,
	}
}

type Delete struct {
	Message string
}

// FlushDB wipes the db
func (hdl *AdminHTTPHandler) WipeDB(c *gin.Context) {
	err := hdl.storage.WipeDB()
	if err != nil {
		hdl.HandleError(c, http.StatusInternalServerError, err)
		return
	}
	res := &Delete{Message: "wiped db ok"}
	c.JSON(http.StatusOK, res)
}
