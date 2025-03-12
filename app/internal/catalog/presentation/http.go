package presentation

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bratushkadan/floral/internal/catalog/api"
	oapi_codegen "github.com/bratushkadan/floral/internal/catalog/presentation/generated"
	"github.com/bratushkadan/floral/internal/catalog/service"
	"github.com/bratushkadan/floral/pkg/xhttp"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=oapi/config.yaml oapi/api.yaml

var _ oapi_codegen.ServerInterface = (*ApiImpl)(nil)

type ApiImpl struct {
	Logger  *zap.Logger
	Service *service.Catalog
}

func (a ApiImpl) CatalogGet(c *gin.Context, params oapi_codegen.CatalogGetParams) {
	res, err := a.Service.Search(c.Request.Context(), service.SearchReq{
		Term:          params.Filter,
		NextPageToken: params.NextPageToken,
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

// Upsert/Delete (product) records from catalog
func (a ApiImpl) CatalogSync(c *gin.Context) {
	var body api.DataStreamProductChangeCdcMessages
	if err := json.NewDecoder(c.Request.Body).Decode(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "bad data stream product change cdc messages body payload", "error": err.Error()})
		return
	}
	if len(body.Messages) == 0 {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "data stream product change cdc messages in body payload length can't be 0"})
		return
	}

	if err := a.Service.Sync(c.Request.Context(), body); err != nil {
		a.Logger.Error("failed to unmarshal CDC pictures string with json list to json", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to sync catalog"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("synced %d items", len(body.Messages))})
}

func (*ApiImpl) ErrorHandlerValidation(c *gin.Context, message string, code int) {
	c.AbortWithStatusJSON(code, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: code, Message: message}))
}
func (*ApiImpl) ErrorHandler(c *gin.Context, err error, code int) {
	c.AbortWithStatusJSON(code, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: code, Message: err.Error()}))
}
