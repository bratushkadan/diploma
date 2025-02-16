package presentation

import (
	"net/http"
	"slices"
	"strings"

	oapi_codegen "github.com/bratushkadan/floral/internal/products/presentation/generated"
	"github.com/bratushkadan/floral/pkg/xhttp"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=oapi/config.yaml oapi/api.yaml

var _ oapi_codegen.ServerInterface = (*ApiImpl)(nil)

type ApiImpl struct {
	Logger *zap.Logger
}

var _ oapi_codegen.ServerInterface = (*ApiImpl)(nil)

func (*ApiImpl) ProductsList(c *gin.Context, params oapi_codegen.ProductsListParams) {
	c.JSON(http.StatusOK, oapi_codegen.ListProductsRes{
		NextPageToken: "",
		Products: []oapi_codegen.ListProductsResProduct{
			{
				Id:         "1",
				Name:       "foo",
				PictureUrl: "https://www.ferra.ru/imgs/2024/05/08/05/6460496/c2150453d059e8999c5f0b211ce334f7c869147c.jpg",
				SellerId:   "3",
			},
		},
	})
}

type authRequiredRecord struct {
	Method string
	Path   string
}

var authRequired = []authRequiredRecord{
	{
		Method: oapi_codegen.ProductsCreateMethod,
		Path:   oapi_codegen.ProductsCreatePath,
	},
	{
		Method: oapi_codegen.ProductsUpdateMethod,
		Path:   oapi_codegen.ProductsUpdatePath,
	},
	{
		Method: oapi_codegen.ProductsDeleteMethod,
		Path:   oapi_codegen.ProductsDeletePath,
	},
}

func checkAuthRequired(c *gin.Context) bool {
	return slices.ContainsFunc(authRequired, func(r authRequiredRecord) bool {
		return c.FullPath() == r.Path && c.Request.Method == r.Method
	})
}

func (a *ApiImpl) AuthMiddleware(c *gin.Context) {
	if checkAuthRequired(c) {
		// do check

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			a.Logger.Info("did not find authorization header value", zap.String("value", c.GetHeader("Authorization")))
			c.AbortWithStatusJSON(http.StatusUnauthorized, oapi_codegen.Error{
				Errors: []oapi_codegen.Err{{Code: 121, Message: "Authorization header must be provided"}},
			})
			return
		}
		token, stripped := strings.CutPrefix(authHeader, "Bearer ")
		if !stripped {
			a.Logger.Info("did not find authorization bearer header", zap.String("value", c.GetHeader("Authorization")))
			c.AbortWithStatusJSON(http.StatusUnauthorized, oapi_codegen.Error{
				Errors: []oapi_codegen.Err{{Code: 122, Message: `authorization header "Bearer " prefix must be provided`}},
			})
			return
		}

		a.Logger.Info("token", zap.String("token", token))
		// process token here
		var err error
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{})
			return
		}
	}

	a.Logger.Info("authorization header", zap.String("value", c.GetHeader("Authorization")))
	c.Next()
}

func (*ApiImpl) ProductsGet(c *gin.Context, id string) {
	c.JSON(http.StatusInternalServerError, oapi_codegen.Error{
		Errors: []oapi_codegen.Err{{Code: 0, Message: "not implemented"}},
	})
}
func (*ApiImpl) ProductsUpdate(c *gin.Context, id string) {
	c.JSON(http.StatusInternalServerError, oapi_codegen.Error{
		Errors: []oapi_codegen.Err{{Code: 0, Message: "not implemented"}},
	})
}
func (*ApiImpl) ProductsCreate(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, oapi_codegen.Error{
		Errors: []oapi_codegen.Err{{Code: 0, Message: "not implemented"}},
	})
}
func (*ApiImpl) ProductsDelete(c *gin.Context, id string) {
	c.JSON(http.StatusInternalServerError, oapi_codegen.Error{
		Errors: []oapi_codegen.Err{{Code: 0, Message: "not implemented"}},
	})
}

func (*ApiImpl) ErrorHandler(c *gin.Context, err error, code int) {
	c.JSON(code, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: code, Message: err.Error()}))
}
