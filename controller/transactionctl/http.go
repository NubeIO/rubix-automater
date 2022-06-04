package transactionctl

import (
	"encoding/json"
	"fmt"
	"github.com/NubeIO/rubix-automater/automater"
	"github.com/NubeIO/rubix-automater/automater/model"
	"github.com/NubeIO/rubix-automater/controller"
	"github.com/NubeIO/rubix-automater/pkg/helpers/apperrors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// TransactionHTTPHandler is an HTTP controller that exposes result endpoints.
type TransactionHTTPHandler struct {
	controller.HTTPHandler
	storage automater.Storage
}

// NewTransactionHTTPHandler creates and returns a new ResultHTTPHandler.
func NewTransactionHTTPHandler(storage automater.Storage) *TransactionHTTPHandler {
	return &TransactionHTTPHandler{
		//transService: resultService,
		storage: storage,
	}
}

// GetTransactions fetches all jobs, optionally filters them by status.
func (hdl *TransactionHTTPHandler) GetTransactions(c *gin.Context) {
	var status string
	value, ok := c.GetQuery("status")
	if ok {
		status = value
	}

	var jobStatus model.JobStatus
	if status == "" {
		jobStatus = model.Undefined
	} else {
		err := json.Unmarshal([]byte("\""+strings.ToUpper(status)+"\""), &jobStatus)
		if err != nil {
			//return nil, err
		}
	}
	fmt.Println(jobStatus)
	//jobStatus = model.Completed

	out, err := hdl.storage.GetTransactions(jobStatus)
	if err != nil {
		switch err.(type) {
		case *apperrors.ResourceValidationErr:
			hdl.HandleError(c, http.StatusBadRequest, err)
			return
		default:
			hdl.HandleError(c, http.StatusInternalServerError, err)
			return
		}
	}
	res := map[string]interface{}{
		"transactions": out,
	}
	c.JSON(http.StatusOK, res)
}
