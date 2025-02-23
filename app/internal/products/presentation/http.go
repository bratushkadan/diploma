package presentation

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	oapi_codegen "github.com/bratushkadan/floral/internal/products/presentation/generated"
	"github.com/bratushkadan/floral/internal/products/service"
	"github.com/bratushkadan/floral/internal/products/store"
	"github.com/bratushkadan/floral/pkg/xhttp"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=oapi/config.yaml oapi/api.yaml

var _ oapi_codegen.ServerInterface = (*ApiImpl)(nil)

type ApiImpl struct {
	Logger          *zap.Logger
	ProductsService *service.Products
	PictureStore    *store.Pictures
}

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

func (a *ApiImpl) ProductsGet(c *gin.Context, id string) {
	parsedId, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: "invalid product id provided"}},
		})
		return
	}

	product, err := a.ProductsService.GetProduct(c.Request.Context(), parsedId)
	if err != nil {
		msg := "failed to retrieve product"
		a.Logger.Error(msg, zap.Error(err))
		c.JSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: msg}},
		})
		return
	}
	if product == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: fmt.Sprintf(`product id="%s" not found`, id)}},
		})
		return
	}

	c.JSON(http.StatusOK, *product)
}
func (*ApiImpl) ProductsUpdate(c *gin.Context, id string) {
	c.JSON(http.StatusInternalServerError, oapi_codegen.Error{
		Errors: []oapi_codegen.Err{{Code: 0, Message: "not implemented"}},
	})
}
func (a *ApiImpl) ProductsCreate(c *gin.Context) {
	var bodyReq oapi_codegen.CreateProductReq
	if err := c.ShouldBindBodyWithJSON(&bodyReq); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "bad request body: " + err.Error()}},
		})
		return
	}

	// accessToken, ok := auth.AccessTokenFromContext(c.Request.Context())
	// if !ok {
	// 	c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
	// 		Errors: []oapi_codegen.Err{{Code: 124, Message: "authentication problems on the server side"}},
	// 	})
	// 	return
	// }

	// _ = accessToken

	// FIXME: token
	product, err := a.ProductsService.CreateProduct(c.Request.Context(), &bodyReq, "dan")
	if err != nil {
		msg := "failed to create product"
		a.Logger.Error(msg, zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: msg}},
		})
		return
	}

	c.JSON(http.StatusOK, product)
}
func (*ApiImpl) ProductsDelete(c *gin.Context, id string) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
		Errors: []oapi_codegen.Err{{Code: 0, Message: "not implemented"}},
	})
}

var allowedExtensions = map[string]struct{}{".jpg": struct{}{}, ".jpeg": struct{}{}, ".png": struct{}{}, ".svg": struct{}{}, ".webp": struct{}{}}

const MiB = 1 << 20
const MaxProductPictureSizeMiB = 2
const MaxProductPictureSize = MaxProductPictureSizeMiB * MiB

func (a *ApiImpl) getPictureS3Path(productId, pictureFilename string) string {
	return productId + "/" + pictureFilename
}

func (a *ApiImpl) ProductsUploadPicture(c *gin.Context, productId string) {
	form, err := c.MultipartForm()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: `bad form data provided`}},
		})
		return
	}
	_ = form
	file, err := c.FormFile("file")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: `bad picture provided in the "file" form field`}},
		})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if _, ok := allowedExtensions[ext]; !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: fmt.Sprintf(`bad picture extension "%s"`, ext)}},
		})
		return
	}

	if file.Size > MaxProductPictureSize {
		c.AbortWithStatusJSON(http.StatusBadRequest, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: fmt.Sprintf(`picture size (%.2f MiB) exceeds max size of %d MiB`, float64(file.Size)/MiB, MaxProductPictureSizeMiB)}},
		})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: "failed to open picture file"}},
		})
		return
	}
	defer func() { _ = f.Close() }()

	picId := uuid.NewString()
	picPath := a.getPictureS3Path(productId, picId+ext)

	pictureUploadRes, err := a.PictureStore.Upload(c.Request.Context(), picPath, f)
	if err != nil {
		a.Logger.Error("failed to upload file to s3", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: "failed to upload file"}},
		})
		return
	}

	c.JSON(http.StatusOK, oapi_codegen.UploadProductPictureRes{
		Id:  picId,
		Url: pictureUploadRes.PictureUrl,
	})
}
func (a *ApiImpl) ProductsDeletePicture(c *gin.Context, productId string, id string) {
	// FIXME: "now pretend that file extension is returned from the record retrieved by using the productId"
	ext := ".png"

	picPath := a.getPictureS3Path(productId, id+ext)

	_, err := a.PictureStore.Delete(c.Request.Context(), picPath)
	if err != nil {
		a.Logger.Error("failed to delete s3 object", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: "failed to delete file"}},
		})
		return
	}

	c.JSON(http.StatusOK, oapi_codegen.DeleteProductPictureRes{Id: id})
}

func (*ApiImpl) ErrorHandlerValidation(c *gin.Context, message string, code int) {
	c.JSON(code, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: code, Message: message}))
}
func (*ApiImpl) ErrorHandler(c *gin.Context, err error, code int) {
	c.JSON(code, xhttp.NewErrorResponse(xhttp.ErrorResponseErr{Code: code, Message: err.Error()}))
}
