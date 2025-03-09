package presentation

import (
	"net/http"

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

func (a *ApiImpl) CatalogGet(c *gin.Context, params oapi_codegen.CatalogGetParams) {
	// params.NextPageToken
	if params.Filter == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "filter must be provided"})
	}

	a.Service
}

func (*ApiImpl) ErrorHandlerValidation(c *gin.Context, message string, code int) {
	c.JSON(code, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: code, Message: message}))
}
func (*ApiImpl) ErrorHandler(c *gin.Context, err error, code int) {
	c.JSON(code, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: code, Message: err.Error()}))
}
