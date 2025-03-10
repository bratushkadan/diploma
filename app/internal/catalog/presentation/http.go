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

func (a ApiImpl) CatalogGet(c *gin.Context, params oapi_codegen.CatalogGetParams) {
	// params.NextPageToken

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

func (*ApiImpl) ErrorHandlerValidation(c *gin.Context, message string, code int) {
	c.AbortWithStatusJSON(code, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: code, Message: message}))
}
func (*ApiImpl) ErrorHandler(c *gin.Context, err error, code int) {
	c.AbortWithStatusJSON(code, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: code, Message: err.Error()}))
}
