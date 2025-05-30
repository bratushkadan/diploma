// Package oapi_codegen provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.4.1 DO NOT EDIT.
package oapi_codegen

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	"github.com/oapi-codegen/runtime"
)

const (
	BearerAuthScopes = "bearerAuth.Scopes"
)

// Err defines model for Err.
type Err struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// FeedbackCreateProductReviewReq defines model for FeedbackCreateProductReviewReq.
type FeedbackCreateProductReviewReq struct {
	Rating float64 `json:"rating"`
	Review string  `json:"review"`
}

// FeedbackCreateProductReviewRes defines model for FeedbackCreateProductReviewRes.
type FeedbackCreateProductReviewRes struct {
	CreatedAt string  `json:"created_at"`
	Id        string  `json:"id"`
	ProductId string  `json:"product_id"`
	Rating    float64 `json:"rating"`
	Review    string  `json:"review"`
	UpdatedAt string  `json:"updated_at"`
	UserId    string  `json:"user_id"`
}

// FeedbackDeleteProductReviewRes defines model for FeedbackDeleteProductReviewRes.
type FeedbackDeleteProductReviewRes struct {
	Id string `json:"id"`
}

// FeedbackGetProductRatingRes defines model for FeedbackGetProductRatingRes.
type FeedbackGetProductRatingRes struct {
	ProductId string `json:"product_id"`
	Rating    int    `json:"rating"`
}

// FeedbackGetProductReviewRes defines model for FeedbackGetProductReviewRes.
type FeedbackGetProductReviewRes struct {
	CreatedAt string  `json:"created_at"`
	Id        string  `json:"id"`
	ProductId string  `json:"product_id"`
	Rating    float64 `json:"rating"`
	Review    string  `json:"review"`
	UpdatedAt string  `json:"updated_at"`
	UserId    string  `json:"user_id"`
}

// FeedbackListProductReviewsRes defines model for FeedbackListProductReviewsRes.
type FeedbackListProductReviewsRes struct {
	NextPageToken *string                               `json:"next_page_token"`
	Reviews       []FeedbackListProductReviewsResReview `json:"reviews"`
}

// FeedbackListProductReviewsResReview defines model for FeedbackListProductReviewsResReview.
type FeedbackListProductReviewsResReview struct {
	CreatedAt string  `json:"created_at"`
	Id        string  `json:"id"`
	ProductId string  `json:"product_id"`
	Rating    float64 `json:"rating"`
	Review    string  `json:"review"`
	UpdatedAt string  `json:"updated_at"`
	UserId    string  `json:"user_id"`
}

// FeedbackUpdateProductReviewReq defines model for FeedbackUpdateProductReviewReq.
type FeedbackUpdateProductReviewReq struct {
	Rating *float64 `json:"rating,omitempty"`
	Review *string  `json:"review,omitempty"`
}

// FeedbackUpdateProductReviewRes defines model for FeedbackUpdateProductReviewRes.
type FeedbackUpdateProductReviewRes struct {
	Id     string   `json:"id"`
	Rating *float64 `json:"rating,omitempty"`
	Review *string  `json:"review,omitempty"`
}

// PrivateFeedbackProcessCompletedOrderReq defines model for PrivateFeedbackProcessCompletedOrderReq.
type PrivateFeedbackProcessCompletedOrderReq struct {
	Messages []PrivateFeedbackProcessCompletedOrderReqMessage `json:"messages"`
}

// PrivateFeedbackProcessCompletedOrderReqMessage defines model for PrivateFeedbackProcessCompletedOrderReqMessage.
type PrivateFeedbackProcessCompletedOrderReqMessage struct {
	OrderId  string                                          `json:"order_id"`
	Products PrivateFeedbackProcessCompletedOrderReqProducts `json:"products"`
}

// PrivateFeedbackProcessCompletedOrderReqProducts defines model for PrivateFeedbackProcessCompletedOrderReqProducts.
type PrivateFeedbackProcessCompletedOrderReqProducts struct {
	Id string `json:"id"`
}

// PrivateFeedbackProcessCompletedOrderRes defines model for PrivateFeedbackProcessCompletedOrderRes.
type PrivateFeedbackProcessCompletedOrderRes = map[string]interface{}

// Error defines model for Error.
type Error struct {
	Errors []Err `json:"errors"`
}

// FeedbackListProductReviewsParams defines parameters for FeedbackListProductReviews.
type FeedbackListProductReviewsParams struct {
	NextPageToken *string `form:"next_page_token,omitempty" json:"next_page_token,omitempty"`
}

// FeedbackProcessCompletedOrderJSONRequestBody defines body for FeedbackProcessCompletedOrder for application/json ContentType.
type FeedbackProcessCompletedOrderJSONRequestBody = PrivateFeedbackProcessCompletedOrderReq

// FeedbackAddProductReviewJSONRequestBody defines body for FeedbackAddProductReview for application/json ContentType.
type FeedbackAddProductReviewJSONRequestBody = FeedbackCreateProductReviewReq

// FeedbackUpdateProductReviewJSONRequestBody defines body for FeedbackUpdateProductReview for application/json ContentType.
type FeedbackUpdateProductReviewJSONRequestBody = FeedbackUpdateProductReviewReq

// Method & Path constants for routes.
// Process completed order contents
const FeedbackProcessCompletedOrderMethod = "POST"
const FeedbackProcessCompletedOrderPath = "/api/private/v1/feedback/orders:process_completed_order"

// Get product rating
const FeedbackGetProductRatingMethod = "GET"
const FeedbackGetProductRatingPath = "/api/v1/feedback/products/:product_id/rating"

// List product reviews
const FeedbackListProductReviewsMethod = "GET"
const FeedbackListProductReviewsPath = "/api/v1/feedback/products/:product_id/reviews"

// Add product review
const FeedbackAddProductReviewMethod = "POST"
const FeedbackAddProductReviewPath = "/api/v1/feedback/products/:product_id/reviews"

// Delete product review
const FeedbackDeleteProductReviewMethod = "DELETE"
const FeedbackDeleteProductReviewPath = "/api/v1/feedback/reviews/:product_id/reviews/:review_id"

// Get product review
const FeedbackGetProductReviewMethod = "GET"
const FeedbackGetProductReviewPath = "/api/v1/feedback/reviews/:product_id/reviews/:review_id"

// Update product review
const FeedbackUpdateProductReviewMethod = "PATCH"
const FeedbackUpdateProductReviewPath = "/api/v1/feedback/reviews/:product_id/reviews/:review_id"

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Process completed order contents
	// (POST /api/private/v1/feedback/orders:process_completed_order)
	FeedbackProcessCompletedOrder(c *gin.Context)
	// Get product rating
	// (GET /api/v1/feedback/products/{product_id}/rating)
	FeedbackGetProductRating(c *gin.Context, productId string)
	// List product reviews
	// (GET /api/v1/feedback/products/{product_id}/reviews)
	FeedbackListProductReviews(c *gin.Context, productId string, params FeedbackListProductReviewsParams)
	// Add product review
	// (POST /api/v1/feedback/products/{product_id}/reviews)
	FeedbackAddProductReview(c *gin.Context, productId string)
	// Delete product review
	// (DELETE /api/v1/feedback/reviews/{product_id}/reviews/{review_id})
	FeedbackDeleteProductReview(c *gin.Context, productId string, reviewId string)
	// Get product review
	// (GET /api/v1/feedback/reviews/{product_id}/reviews/{review_id})
	FeedbackGetProductReview(c *gin.Context, productId string, reviewId string)
	// Update product review
	// (PATCH /api/v1/feedback/reviews/{product_id}/reviews/{review_id})
	FeedbackUpdateProductReview(c *gin.Context, productId string, reviewId string)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandler       func(*gin.Context, error, int)
}

type MiddlewareFunc func(c *gin.Context)

// FeedbackProcessCompletedOrder operation middleware
func (siw *ServerInterfaceWrapper) FeedbackProcessCompletedOrder(c *gin.Context) {

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.FeedbackProcessCompletedOrder(c)
}

// FeedbackGetProductRating operation middleware
func (siw *ServerInterfaceWrapper) FeedbackGetProductRating(c *gin.Context) {

	var err error

	// ------------- Path parameter "product_id" -------------
	var productId string

	err = runtime.BindStyledParameterWithOptions("simple", "product_id", c.Param("product_id"), &productId, runtime.BindStyledParameterOptions{Explode: false, Required: true})
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter product_id: %w", err), http.StatusBadRequest)
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.FeedbackGetProductRating(c, productId)
}

// FeedbackListProductReviews operation middleware
func (siw *ServerInterfaceWrapper) FeedbackListProductReviews(c *gin.Context) {

	var err error

	// ------------- Path parameter "product_id" -------------
	var productId string

	err = runtime.BindStyledParameterWithOptions("simple", "product_id", c.Param("product_id"), &productId, runtime.BindStyledParameterOptions{Explode: false, Required: true})
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter product_id: %w", err), http.StatusBadRequest)
		return
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params FeedbackListProductReviewsParams

	// ------------- Optional query parameter "next_page_token" -------------

	err = runtime.BindQueryParameter("form", true, false, "next_page_token", c.Request.URL.Query(), &params.NextPageToken)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter next_page_token: %w", err), http.StatusBadRequest)
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.FeedbackListProductReviews(c, productId, params)
}

// FeedbackAddProductReview operation middleware
func (siw *ServerInterfaceWrapper) FeedbackAddProductReview(c *gin.Context) {

	var err error

	// ------------- Path parameter "product_id" -------------
	var productId string

	err = runtime.BindStyledParameterWithOptions("simple", "product_id", c.Param("product_id"), &productId, runtime.BindStyledParameterOptions{Explode: false, Required: true})
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter product_id: %w", err), http.StatusBadRequest)
		return
	}

	c.Set(BearerAuthScopes, []string{})

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.FeedbackAddProductReview(c, productId)
}

// FeedbackDeleteProductReview operation middleware
func (siw *ServerInterfaceWrapper) FeedbackDeleteProductReview(c *gin.Context) {

	var err error

	// ------------- Path parameter "product_id" -------------
	var productId string

	err = runtime.BindStyledParameterWithOptions("simple", "product_id", c.Param("product_id"), &productId, runtime.BindStyledParameterOptions{Explode: false, Required: true})
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter product_id: %w", err), http.StatusBadRequest)
		return
	}

	// ------------- Path parameter "review_id" -------------
	var reviewId string

	err = runtime.BindStyledParameterWithOptions("simple", "review_id", c.Param("review_id"), &reviewId, runtime.BindStyledParameterOptions{Explode: false, Required: true})
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter review_id: %w", err), http.StatusBadRequest)
		return
	}

	c.Set(BearerAuthScopes, []string{})

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.FeedbackDeleteProductReview(c, productId, reviewId)
}

// FeedbackGetProductReview operation middleware
func (siw *ServerInterfaceWrapper) FeedbackGetProductReview(c *gin.Context) {

	var err error

	// ------------- Path parameter "product_id" -------------
	var productId string

	err = runtime.BindStyledParameterWithOptions("simple", "product_id", c.Param("product_id"), &productId, runtime.BindStyledParameterOptions{Explode: false, Required: true})
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter product_id: %w", err), http.StatusBadRequest)
		return
	}

	// ------------- Path parameter "review_id" -------------
	var reviewId string

	err = runtime.BindStyledParameterWithOptions("simple", "review_id", c.Param("review_id"), &reviewId, runtime.BindStyledParameterOptions{Explode: false, Required: true})
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter review_id: %w", err), http.StatusBadRequest)
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.FeedbackGetProductReview(c, productId, reviewId)
}

// FeedbackUpdateProductReview operation middleware
func (siw *ServerInterfaceWrapper) FeedbackUpdateProductReview(c *gin.Context) {

	var err error

	// ------------- Path parameter "product_id" -------------
	var productId string

	err = runtime.BindStyledParameterWithOptions("simple", "product_id", c.Param("product_id"), &productId, runtime.BindStyledParameterOptions{Explode: false, Required: true})
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter product_id: %w", err), http.StatusBadRequest)
		return
	}

	// ------------- Path parameter "review_id" -------------
	var reviewId string

	err = runtime.BindStyledParameterWithOptions("simple", "review_id", c.Param("review_id"), &reviewId, runtime.BindStyledParameterOptions{Explode: false, Required: true})
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter review_id: %w", err), http.StatusBadRequest)
		return
	}

	c.Set(BearerAuthScopes, []string{})

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.FeedbackUpdateProductReview(c, productId, reviewId)
}

// GinServerOptions provides options for the Gin server.
type GinServerOptions struct {
	BaseURL      string
	Middlewares  []MiddlewareFunc
	ErrorHandler func(*gin.Context, error, int)
}

// RegisterHandlers creates http.Handler with routing matching OpenAPI spec.
func RegisterHandlers(router gin.IRouter, si ServerInterface) {
	RegisterHandlersWithOptions(router, si, GinServerOptions{})
}

// RegisterHandlersWithOptions creates http.Handler with additional options
func RegisterHandlersWithOptions(router gin.IRouter, si ServerInterface, options GinServerOptions) {
	errorHandler := options.ErrorHandler
	if errorHandler == nil {
		errorHandler = func(c *gin.Context, err error, statusCode int) {
			c.JSON(statusCode, gin.H{"msg": err.Error()})
		}
	}

	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandler:       errorHandler,
	}

	router.POST(options.BaseURL+"/api/private/v1/feedback/orders:process_completed_order", wrapper.FeedbackProcessCompletedOrder)
	router.GET(options.BaseURL+"/api/v1/feedback/products/:product_id/rating", wrapper.FeedbackGetProductRating)
	router.GET(options.BaseURL+"/api/v1/feedback/products/:product_id/reviews", wrapper.FeedbackListProductReviews)
	router.POST(options.BaseURL+"/api/v1/feedback/products/:product_id/reviews", wrapper.FeedbackAddProductReview)
	router.DELETE(options.BaseURL+"/api/v1/feedback/reviews/:product_id/reviews/:review_id", wrapper.FeedbackDeleteProductReview)
	router.GET(options.BaseURL+"/api/v1/feedback/reviews/:product_id/reviews/:review_id", wrapper.FeedbackGetProductReview)
	router.PATCH(options.BaseURL+"/api/v1/feedback/reviews/:product_id/reviews/:review_id", wrapper.FeedbackUpdateProductReview)
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/+xab2/bNhP/KgKf58XzAHLkttg66F3WZUWxFTXSDhhQBMaZPNtsJVEhKTee4e8+8I8k",
	"25JsObGTIuirKNLd8Xj3u3+kV4SKNBcZZlqReEUkqlxkCu0/V1IKaR6oyDRm2jxCniecguYii74okZl3",
	"is4xBfuVMW4+QTKSIkepuZE0hURhSPKNVyuCRrh94hpT+/BfiVMSk/9EtU6Rk62iKynJOiR6mSOJCUgJ",
	"S7Jeh0TibcElMhJ/LkXeVGRi8gWpJmtDyFBRyXOjHYkdqRXgF/D7PXITVDA0f/16PNM4Q6toikrBbPOj",
	"0pJns4bSVkRN31Q+JL8jsgnQr28kgsaRFKyg+hoXHL9d4+2RKkvQRo14RaZCpqBJTJgoJolRwq+cFenE",
	"7ULaRQ5vwgutGI7ehTrW8FYIG4NuUS4knLW+zt2i447PpzFNSIqc7VOuUCjbVdixKmdkS+eaNWxYPNw0",
	"yZYK+3zxGyb4YF/03Mo+Pd6iLpWw+zpeiQe4NoU7nhYpiX8KScoz9zwMG0G9s6Utx3j5Pff4A/RPCfo/",
	"udr2hDreFRne6XEOMxxr8RVtFcyKJAFjt1jLAsMWS7vFepe8vfq6p4MlsVwzbGh8tIWuK9//gOxjQ/Yv",
	"S/a9VP7j1DxJOQnP07d0lKWR5AvQWG5rJAVFpd6INDflkn2QDOXx9vctXv8E0FON9751PJQLqvUbWw7J",
	"3UDDTLmyZhcdQ87JTX9bvK/73SNMIgz3eH8WOJWZRqW4jkKutuK20uwcxhpt7OwRGq2Ha6w2Fj4o1AxV",
	"SAvJ9fKjcZHjniBIlJeFntutmCFsjsBQmtoEqZH898B8FpL/Y+fLOrAh53/g0s1xPJsKqw3XptaSKyrS",
	"4HL0joRkgVK58W548eJiaDAkcsyMVjF5dTG8GJo8DXpuFYog55HXPFq8iChIHdMEQQ78vGvJ7gaeZmDl",
	"mNq+DtuZ82KScDW/B/vUmz+ysFNx7twwpqUfxvaDbXSFsuVqe571fgsqhsAyBKUqwf+UCPQcdGAgHsxB",
	"BRDkKFOujMkCLYIEYYFBqUkwFdKQOKT+n1hTSuuXd4zEZC9giMMkKv2rYMujThBOEOgGJzYqNo4yXg6H",
	"j6yGajt2KAM/YKCB2M9TKBLdtWS1h+iqPrAo0hTksofTTQD5CC3damO+HyatrHgCms4HFDKKyaDIcuBs",
	"4EDaH91OkpdxP2YfEAMfYsgGJuIGuVA2ed5bnkSFcoFssFls+gkqOWIv4x6cRdaP1yeYaOXr0zra3HgP",
	"qmhVt6frXRYNiZhtv6zyUanqloCo7sVm2JKM3qIuE0dQ9b/t6WN38LcJWkKK2oLk867oUqyt0baImHxe",
	"l5Dtobyqi24iq8N7t4benDFd7Dvj6E4R3nBBDstEAHtwrmj1STM7mOSwpAaBM9D4DZYDe/zhPFeaBHjm",
	"uzZCTK2XC05xDJSKInNDG4Ev+Br56/zrz0n+cji9/eX1q2ldzy3mZeIqnJenyLq5+AISzkC7U2j/D15v",
	"lhYbJusjoFsP4a3YNaNvbahqem5Hb3NOfiL8hr6jui1QLmtpuzP/U4dA+8nLviBwVCeLgg7vPps4CDva",
	"w0vGdrbdielLxrZc9JQZ+fQt5IGrlDN3jgeuQM4fCH4ss27cHMg+3xiD13HSipdnXS28gVuLRbRyD755",
	"IszemjSjzN2m9A20lruXp6se2+s4zTuXqazx3fVYHfdZLYHlKBsoP3NcdSHkGRWgwxPB/rDYvSb7EROn",
	"mzv6V5rzzB3PEO45aDpvAt7dhPTFfMu9yfOH/fmau47bskdq7jouwQ6HXOEx8zgtXhdCn2OXZ8/8oir2",
	"VLSqnpsHYp64PKbs+BKtyhuiHfaNg8SWt3sO4trPK3JOdSFRHUMbrRrCC2V0jqHQc8y0ATi2ElCRTblM",
	"x5gCT9yp5Lb1qf/RXop6LpgyKP7w8ZO9MOMzkyDKxLgr2I48l5SiUp/8Tyb2UBnU7KH4iEmCch+dxDwB",
	"itc4lajm1YrrCuCNEbk2DRdZ4BFcZzpjOtLMj9WJfoOhAkKT6Y07c23ylIexbSxSt9FL3UJsryBaVPKg",
	"7txFNWY2OKtssL5Z/xsAAP//srlMmTQqAAA=",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
