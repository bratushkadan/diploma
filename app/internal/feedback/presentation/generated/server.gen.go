// Package oapi_codegen provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.4.1 DO NOT EDIT.
package oapi_codegen

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
)

// Err defines model for Err.
type Err struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
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
type PrivateFeedbackProcessCompletedOrderRes struct {
	Message *string `json:"message,omitempty"`
}

// Error defines model for Error.
type Error struct {
	Errors []Err `json:"errors"`
}

// FeedbackProcessCompletedOrderJSONRequestBody defines body for FeedbackProcessCompletedOrder for application/json ContentType.
type FeedbackProcessCompletedOrderJSONRequestBody = PrivateFeedbackProcessCompletedOrderReq

// Method & Path constants for routes.
// Process completed order contents
const FeedbackProcessCompletedOrderMethod = "POST"
const FeedbackProcessCompletedOrderPath = "/api/private/v1/feedback/orders:process_completed_order"

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Process completed order contents
	// (POST /api/private/v1/feedback/orders:process_completed_order)
	FeedbackProcessCompletedOrder(c *gin.Context)
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
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/7xWTW/bRhD9K8S0hxagTKW98eYGLhAUQYQ4hwKFYYyXI3ETkruZHSpRBf73YpYfUiza",
	"lYIkJ9O7b2bezscb7cG42ruGGgmQ74EpeNcEiv/cMDvWD+MaoUb0E72vrEGxrsneB9foWTAl1Rhvi8Lq",
	"FVYrdp5YrHpaYxUoBX90tAdS5/HLCtXx42emNeTwU3bglPW+Q3bDDF0KsvMEOSAz7qDrUmD62FqmAvJ/",
	"Rpd3E8w9vCcj0CmwoGDYemUHeQ+NDoYAw3svfIRxBenfIZ5thDYUidYUAm6OL4OwbTYnpKOLA/6UfAor",
	"tlsU+pOoeEDzYcXOUAgvXe0rEirecEH8lj5eyH2IeH4JzqTxenjJ/1Vrin/y5BQ+LwQ3QVG+D3qP3sLd",
	"+bl4fUj/BSlxan1vi5mqRWjRGvlWaVqN7h7nZYqTQht6OumB2fdI1uroZRdkazZPjx7zfRiHr2v1ebrn",
	"klOtINOyld2tlrp3/UDIxNetlDElqi0lYUEMKTRYq+e/F3rt2P4bZRMOc+HtX7Tr5ck2axf5Wan07sa4",
	"OrlevYIUtsShV63l1Yurpfai89Qoqxx+v1peLSEFj1JGQhl6mw3Ms+2LzCBLbipCXgwyHmGfFwNmEf0I",
	"t9Sl88a+fahsKL/CfD2UMYvtG3Lfl/PejPW8jxfq0LsQ98uXMj3UP5kMkmiQjFSSX4JLpERJdFSSEkOC",
	"iSeubdCUJeKSinBLycgkWTtWSN/xv0JMJce6vCogh2cbD/repiB/uGJ30WL8BoKhfRKn62hD/7Zc/mAa",
	"YW6bjgKSFCgI8XqNbSVPhZzekN0c9nBb18i7M4quAzRM6FjWqB3n9eQorzlTIN7S+d08WbbNebbDAGX7",
	"Qce7zLsQNUuLdwYq2w8x9fixiWDlNl8eHq2omdNnvM1CMm+NtEzhEmy2P3Gu7wpZjq2U1Ig2J80CjGvW",
	"lut7qtFWfWp3RrO6QaFPuFuY4QdjTVK6ImirvLl9F7ej3aj0xuTPOGZCoWujXfXOfaBmPvyIcm0cpacQ",
	"t1RVxM/hmHyFht7SmimUU8Ru6tvHMnd9SI2qlvaWNXTYIJo66NL9U2N3YjA1wqnRy75xTm3GjpozYZnD",
	"s8yAo07MUOpXwNOvSJi2lj7NWE5D3t11/wUAAP//5RedBLAMAAA=",
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
