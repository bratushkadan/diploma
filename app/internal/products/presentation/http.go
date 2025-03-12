package presentation

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"slices"
	"strings"

	oapi_codegen "github.com/bratushkadan/floral/internal/products/presentation/generated"
	"github.com/bratushkadan/floral/internal/products/service"
	"github.com/bratushkadan/floral/internal/products/store"
	"github.com/bratushkadan/floral/pkg/shared/api"
	"github.com/bratushkadan/floral/pkg/xhttp"
	"github.com/bratushkadan/floral/pkg/xhttp/gin/middleware/auth"
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

type productsListFilter struct {
	ProductIds []uuid.UUID
	SellerId   *string
	InStock    *bool
}

func parseProductsListFilter(filter string) (productsListFilter, error) {
	f := productsListFilter{}
	for _, cond := range strings.Split(filter, "&") {
		pair := strings.Split(cond, "=")
		if len(pair) < 2 {
			return productsListFilter{}, fmt.Errorf(`invalid filter condition with key "%s" provided: condition must be a key=value(s) pair`, pair[0])
		}
		switch key, val := pair[0], pair[1]; key {
		case "seller.id":
			f.SellerId = &val
		case "in_stock":
			if val == "*" {
			} else if val == "true" {
				f.InStock = ptr(true)
			} else if val == "false" {
				f.InStock = ptr(false)
			} else {
				return productsListFilter{}, errors.New(`invalid filter condition "in_stock" provided: condition's value must be one of ["*", "true", "false"]`)
			}
		}
	}

	return f, nil
}

func (a *ApiImpl) ProductsList(c *gin.Context, params oapi_codegen.ProductsListParams) {
	if params.NextPageToken == nil && params.Filter == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: `either "filter" or "nextPageToken" query parameter must be specified`}},
		})
		return
	}
	if params.NextPageToken != nil && params.Filter != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: `both "filter" and "nextPageToken" can't be specified`}},
		})
		return
	}
	listProductsReq := service.ListProductsReq{
		NextPageToken: params.NextPageToken,
	}
	if params.NextPageToken == nil {
		filter, err := parseProductsListFilter(*params.Filter)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, oapi_codegen.Error{
				Errors: []oapi_codegen.Err{{Code: 0, Message: fmt.Sprintf("invalid request filter: %v", err)}},
			})
			return
		}
		listProductsReq.Filter = service.ListProductsReqFilter{
			PageSize: params.MaxPageSize,
		}
		if filter.SellerId != nil {
			listProductsReq.Filter.SellerId = filter.SellerId
		}
		if filter.InStock != nil {
			listProductsReq.Filter.InStock = filter.InStock
		}
	}

	res, err := a.ProductsService.ListProducts(c.Request.Context(), listProductsReq)
	if err != nil {
		if errors.Is(err, service.ErrInvalidListProductsPageSize) {
			c.AbortWithStatusJSON(http.StatusBadRequest, oapi_codegen.Error{
				Errors: []oapi_codegen.Err{{Code: 0, Message: err.Error()}},
			})
			return
		}
		if errors.Is(err, service.ErrInvalidListProductsNextPageToken) {
			c.AbortWithStatusJSON(http.StatusBadRequest, oapi_codegen.Error{
				Errors: []oapi_codegen.Err{{Code: 0, Message: err.Error()}},
			})
			return
		}
		msg := "failed to list products"
		a.Logger.Error(msg, zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: msg}},
		})
		return
	}

	c.JSON(http.StatusOK, res)
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
func (a *ApiImpl) ProductsUpdate(c *gin.Context, id string) {
	accessToken, ok := auth.AccessTokenFromContext(c.Request.Context())
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "authentication problems on the server side"}},
		})
		return
	}

	if !slices.Contains([]string{api.SubjectTypeSeller, api.SubjectTypeAdmin}, accessToken.SubjectType) {
		c.AbortWithStatusJSON(http.StatusForbidden, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "permission denied"}},
		})
		return
	}

	sellerId := accessToken.SubjectId

	parsedProductId, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: "invalid product id provided"}},
		})
		return
	}

	product, err := a.ProductsService.GetProduct(c.Request.Context(), parsedProductId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: "failed to retrieve product data for image uploading"}},
		})
	}
	if product == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: fmt.Sprintf(`product id="%s" not found`, id)}},
		})
		return
	}
	if product.SellerId != sellerId && accessToken.SubjectType != api.SubjectTypeAdmin {
		c.AbortWithStatusJSON(http.StatusForbidden, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: "permission denied"}},
		})
		return
	}

	var bodyReq oapi_codegen.UpdateProductReq
	if err := c.ShouldBindBodyWithJSON(&bodyReq); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "bad request body: " + err.Error()}},
		})
		return
	}

	var stockDelta *int32
	if bodyReq.StockDelta != nil {
		stockDelta = ptr(int32(*bodyReq.StockDelta))
	}
	var metadata map[string]any
	if bodyReq.Metadata != nil {
		metadata = *bodyReq.Metadata
	}
	res, err := a.ProductsService.UpdateProduct(c.Request.Context(), service.UpdateProductReq{
		Id:          parsedProductId,
		Name:        bodyReq.Name,
		Description: bodyReq.Description,
		Metadata:    metadata,
		StockDelta:  stockDelta,
		Price:       bodyReq.Price,
	})
	if err != nil {
		if errors.Is(err, service.ErrInsufficientStock) {
			c.AbortWithStatusJSON(http.StatusPreconditionFailed, oapi_codegen.Error{
				Errors: []oapi_codegen.Err{{Code: 0, Message: err.Error()}},
			})
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: "failed to update product"}},
		})
		return
	}

	c.JSON(http.StatusOK, res)
}
func (a *ApiImpl) ProductsCreate(c *gin.Context) {
	accessToken, ok := auth.AccessTokenFromContext(c.Request.Context())
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "authentication problems on the server side"}},
		})
		return
	}

	if !slices.Contains([]string{api.SubjectTypeSeller, api.SubjectTypeAdmin}, accessToken.SubjectType) {
		c.AbortWithStatusJSON(http.StatusForbidden, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "permission denied to upload product picture"}},
		})
		return
	}

	sellerId := accessToken.SubjectId

	var bodyReq oapi_codegen.CreateProductReq
	if err := c.ShouldBindBodyWithJSON(&bodyReq); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "bad request body: " + err.Error()}},
		})
		return
	}

	product, err := a.ProductsService.CreateProduct(c.Request.Context(), &bodyReq, sellerId)
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
func (a *ApiImpl) ProductsDelete(c *gin.Context, id string) {
	accessToken, ok := auth.AccessTokenFromContext(c.Request.Context())
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "authentication problems on the server side"}},
		})
		return
	}

	if !slices.Contains([]string{api.SubjectTypeSeller, api.SubjectTypeAdmin}, accessToken.SubjectType) {
		c.AbortWithStatusJSON(http.StatusForbidden, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "permission denied"}},
		})
		return
	}

	sellerId := accessToken.SubjectId

	parsedProductId, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: "invalid product id provided"}},
		})
		return
	}

	product, err := a.ProductsService.GetProduct(c.Request.Context(), parsedProductId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: "failed to retrieve product for deletion"}},
		})
	}
	if product == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: fmt.Sprintf(`product id="%s" not found`, id)}},
		})
		return
	}
	if product.SellerId != sellerId && accessToken.SubjectType != api.SubjectTypeAdmin {
		c.AbortWithStatusJSON(http.StatusForbidden, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: "permission denied"}},
		})
		return
	}

	res, err := a.ProductsService.DeleteProduct(c.Request.Context(), parsedProductId)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, oapi_codegen.Error{
				Errors: []oapi_codegen.Err{{Code: 0, Message: err.Error()}},
			})
			return
		}
		msg := "failed to delete product"
		a.Logger.Error(msg, zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: msg}},
		})
		return
	}

	c.JSON(http.StatusOK, res)
}

var allowedExtensions = map[string]struct{}{".jpg": struct{}{}, ".jpeg": struct{}{}, ".png": struct{}{}, ".svg": struct{}{}, ".webp": struct{}{}}

const MiB = 1 << 20
const MaxProductPictureSizeMiB = 2
const MaxProductPictureSize = MaxProductPictureSizeMiB * MiB

func (a *ApiImpl) getPictureS3Path(productId, pictureFilename string) string {
	return productId + "/" + pictureFilename
}

const productPicturesLimitCount = 3

func (a *ApiImpl) ProductsUploadPicture(c *gin.Context, productId string) {
	accessToken, ok := auth.AccessTokenFromContext(c.Request.Context())
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "authentication problems on the server side"}},
		})
		return
	}

	if !slices.Contains([]string{api.SubjectTypeSeller, api.SubjectTypeAdmin}, accessToken.SubjectType) {
		c.AbortWithStatusJSON(http.StatusForbidden, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "permission denied to upload product picture"}},
		})
		return
	}

	sellerId := accessToken.SubjectId

	parsedProductId, err := uuid.Parse(productId)
	if err != nil {
		c.JSON(http.StatusBadRequest, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: "invalid product id provided"}},
		})
		return
	}

	product, err := a.ProductsService.GetProduct(c.Request.Context(), parsedProductId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: "failed to retrieve product data for image uploading"}},
		})
	}
	if product == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: "product not found"}},
		})
		return
	}
	if product.SellerId != sellerId && accessToken.SubjectType != api.SubjectTypeAdmin {
		c.AbortWithStatusJSON(http.StatusForbidden, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: "permission denied"}},
		})
		return
	}
	if len(product.Pictures) == productPicturesLimitCount {
		c.AbortWithStatusJSON(http.StatusPreconditionFailed, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: fmt.Sprintf("max amount of pictures of %d is reached, try again after deleting another picture", productPicturesLimitCount)}},
		})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: "bad form data provided"}},
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
			Errors: []oapi_codegen.Err{{Code: 0, Message: fmt.Sprintf("picture size (%.2f MiB) exceeds max size of %d MiB", float64(file.Size)/MiB, MaxProductPictureSizeMiB)}},
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

	// data, err := io.ReadAll(f)
	// if err != nil {
	// 	c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
	// 		Errors: []oapi_codegen.Err{{Code: 0, Message: "failed to read form picture file"}},
	// 	})
	// 	return
	// }

	// if err := os.WriteFile("pic.jpg", data, 0775); err != nil {
	// 	c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
	// 		Errors: []oapi_codegen.Err{{Code: 0, Message: "failed to write form picture file"}},
	// 	})
	// 	return
	// }
	// pictureUploadRes, err := a.PictureStore.Upload(c.Request.Context(), picPath, bytes.NewReader(data))

	// data, err = os.ReadFile("/Users/bratushkadan/Downloads/479dd76f0e3e65ecf0c2120d54bc9a27.jpg")
	// if err != nil {
	// 	c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
	// 		Errors: []oapi_codegen.Err{{Code: 0, Message: "failed to read file from downloads"}},
	// 	})
	// 	return
	// }
	// pictureUploadRes, err := a.PictureStore.Upload(c.Request.Context(), picPath, bytes.NewReader(data))
	pictureUploadRes, err := a.PictureStore.Upload(c.Request.Context(), picPath, f)
	if err != nil {
		a.Logger.Error("failed to upload file to s3", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: "failed to upload file"}},
		})
		return
	}

	pictures := make([]store.UpsertProductDTOOutputPicture, 0, len(product.Pictures)+1)
	for _, p := range product.Pictures {
		pictures = append(pictures, store.UpsertProductDTOOutputPicture{Id: p.Id, Url: p.Url})
	}
	pictures = append(pictures, store.UpsertProductDTOOutputPicture{Id: picId, Url: pictureUploadRes.PictureUrl})

	_, err = a.ProductsService.UpdateProduct(c.Request.Context(), service.UpdateProductReq{
		Id:       parsedProductId,
		Pictures: pictures,
	})
	if err != nil {
		msg := "failed to save picture information to product"
		a.Logger.Error(msg, zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: msg}},
		})
		return
	}

	c.JSON(http.StatusOK, oapi_codegen.UploadProductPictureRes{
		Id:  picId,
		Url: pictureUploadRes.PictureUrl,
	})
}

func (a *ApiImpl) ProductsDeletePicture(c *gin.Context, productId string, id string) {
	accessToken, ok := auth.AccessTokenFromContext(c.Request.Context())
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "authentication problems on the server side"}},
		})
		return
	}

	if !slices.Contains([]string{api.SubjectTypeSeller, api.SubjectTypeAdmin}, accessToken.SubjectType) {
		c.AbortWithStatusJSON(http.StatusForbidden, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 124, Message: "permission denied to upload product picture"}},
		})
		return
	}

	sellerId := accessToken.SubjectId

	parsedProductId, err := uuid.Parse(productId)
	if err != nil {
		c.JSON(http.StatusBadRequest, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: "invalid product id provided"}},
		})
		return
	}

	product, err := a.ProductsService.GetProduct(c.Request.Context(), parsedProductId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: "failed to retrieve product data for image uploading"}},
		})
	}
	if product == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: "product not found"}},
		})
		return
	}
	if product.SellerId != sellerId && accessToken.SubjectType != api.SubjectTypeAdmin {
		c.AbortWithStatusJSON(http.StatusForbidden, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: "permission denied"}},
		})
		return
	}

	var ext *string
	for i, pic := range product.Pictures {
		if pic.Id == id {
			pathWithoutExt := a.getPictureS3Path(productId, id)
			s := strings.Split(pic.Url, pathWithoutExt)
			if len(s) < 2 {
				a.Logger.Error("failed to parse extension from s3 image url, image without extension might have been added to the product",
					zap.String("product_id", productId),
					zap.String("picture_id", id),
					zap.String("s3_url", pic.Url),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
					Errors: []oapi_codegen.Err{{Code: 0, Message: fmt.Sprintf(`failed to process picture "%s" deletion`, id)}},
				})
				return
			}
			ext = &s[1]

			product.Pictures = removed(product.Pictures, i)
			break
		}
	}
	if ext == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: fmt.Sprintf(`product picture id="%s" not found`, id)}},
		})
		return
	}

	var pics = make([]store.UpsertProductDTOOutputPicture, 0, len(product.Pictures))
	for _, p := range product.Pictures {
		pics = append(pics, store.UpsertProductDTOOutputPicture{
			Id:  p.Id,
			Url: p.Url,
		})
	}

	_, err = a.ProductsService.UpdateProduct(c.Request.Context(), service.UpdateProductReq{
		Id:       parsedProductId,
		Pictures: pics,
	})
	if err != nil {
		msg := "failed to delete product image"
		a.Logger.Error("", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, oapi_codegen.Error{
			Errors: []oapi_codegen.Err{{Code: 0, Message: msg}},
		})
		return
	}

	picPath := a.getPictureS3Path(productId, id+*ext)

	_, err = a.PictureStore.Delete(c.Request.Context(), picPath)
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

func removed[T any](s []T, idx int) []T {
	c := make([]T, len(s))
	copy(c, s)
	return append(c[:idx], c[idx+1:]...)
}
func ptr[T any](v T) *T {
	return &v
}
