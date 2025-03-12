package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/bratushkadan/floral/internal/catalog/api"
	oapi_codegen "github.com/bratushkadan/floral/internal/catalog/presentation/generated"
	"github.com/bratushkadan/floral/internal/catalog/store"
	"go.uber.org/zap"
)

type Catalog struct {
	logger *zap.Logger
	store  *store.Store
}

func NewCatalog(store *store.Store, logger *zap.Logger) *Catalog {
	return &Catalog{
		logger: logger,
		store:  store,
	}
}

func (c *Catalog) Get(ctx context.Context, nextPageToken *string) (oapi_codegen.CatalogGetRes, error) {
	out, err := c.store.Search(ctx, store.SearchDTOInput{
		NextPageToken: nextPageToken,
	})
	if err != nil {
		return oapi_codegen.CatalogGetRes{}, err
	}

	res := oapi_codegen.CatalogGetRes{
		Products: make([]oapi_codegen.CatalogGetResProduct, 0, len(out.Products)),
	}

	for _, p := range out.Products {
		res.Products = append(res.Products, oapi_codegen.CatalogGetResProduct{
			Id:      p.Id,
			Name:    p.Name,
			Picture: p.Picture,
			Price:   p.Price,
		})
	}

	res.NextPageToken = out.NextPageToken

	return res, nil
}

type SearchReq struct {
	Term          *string
	NextPageToken *string
}

func (c *Catalog) Search(ctx context.Context, req SearchReq) (oapi_codegen.CatalogGetRes, error) {
	out, err := c.store.Search(ctx, store.SearchDTOInput{
		NextPageToken: req.NextPageToken,
		Term:          req.Term,
	})
	if err != nil {
		return oapi_codegen.CatalogGetRes{}, err
	}

	res := oapi_codegen.CatalogGetRes{
		Products:      make([]oapi_codegen.CatalogGetResProduct, 0, len(out.Products)),
		NextPageToken: out.NextPageToken,
	}

	for _, p := range out.Products {
		res.Products = append(res.Products, oapi_codegen.CatalogGetResProduct{
			Id:      p.Id,
			Name:    p.Name,
			Picture: p.Picture,
			Price:   p.Price,
		})
	}

	return res, nil
}

func (c *Catalog) Sync(ctx context.Context, body api.DataStreamProductChangeCdcMessages) error {
	var blkBuf bytes.Buffer
	blkBuf.WriteByte('\n')
	for _, record := range body.Messages {
		switch record.Payload.Operation {
		case api.CdcOperationUpsert:
			var pictures []api.ProductsChangePicture
			if err := json.Unmarshal([]byte(*record.Payload.After.PicturesJsonListStr), &pictures); err != nil {
				return err
			}
			data, err := base64.StdEncoding.DecodeString(record.Payload.After.Id)
			if err != nil {
				msg := `failed to decode base64 encoded bytes field "id"`
				c.logger.Error(msg, zap.Error(err))
				return fmt.Errorf("%s: %v", msg, err)
			}
			uuidId := string(data)

			isDeleted := record.Payload.After.DeletedAtUnixMs != nil
			isOutOfStock := *record.Payload.After.Stock == 0
			if isDeleted || isOutOfStock {
				bulkItem, err := newBulkProductDelete(api.ProductChange{Id: uuidId})
				if err != nil {
					msg := "failed to prepare bulk delete item"
					c.logger.Error(msg, zap.Error(err))
					return fmt.Errorf("%s: %v", msg, err)
				}
				blkBuf.WriteString(bulkItem)
			} else {
				bulkItem, err := newBulkProductUpsert(api.ProductChange{
					Id:          uuidId,
					Name:        *record.Payload.After.Name,
					Description: *record.Payload.After.Description,
					Price:       *record.Payload.After.Price,
					Stock:       *record.Payload.After.Stock,
					Pictures:    pictures,
				})
				if err != nil {
					msg := "failed to prepare bulk upsert item"
					c.logger.Error(msg, zap.Error(err))
					return fmt.Errorf("%s: %v", msg, err)
				}
				blkBuf.WriteString(bulkItem)
			}
		case api.CdcOperationDelete:
			data, err := base64.StdEncoding.DecodeString(record.Payload.Before.Id)
			if err != nil {
				msg := `failed to decode base64 encoded bytes field "id"`
				c.logger.Error(msg, zap.Error(err))
				return fmt.Errorf("failed to prepare bulk delete item: %v", err)
			}
			uuidId := string(data)
			bulkItemDel, err := newBulkProductDelete(api.ProductChange{Id: uuidId})
			if err != nil {
				return fmt.Errorf("failed to prepare bulk delete item: %v", err)
			}
			blkBuf.WriteString(bulkItemDel)
		default:
			return fmt.Errorf("unkown CDC operation type %s", record.Payload.Operation)
		}
		blkBuf.WriteByte('\n')
	}

	fmt.Println(blkBuf.String())

	blk, err := c.store.Sync(ctx, &blkBuf)
	if err != nil {
		c.logger.Error("failed to sync catalog", zap.Error(err))
		return err
	}
	if blk.StatusCode > 399 {
		data, err := io.ReadAll(blk.Body)
		if err != nil {
			c.logger.Error("failed read OpenSearch bulk response", zap.Int("status", blk.StatusCode), zap.ByteString("response_body", data))
			return fmt.Errorf("failed read OpenSearch bulk response: %v", err)
		}
		c.logger.Error("failed to perform bulk operation in OpenSearch", zap.Int("status", blk.StatusCode), zap.ByteString("response_body", data))
		return fmt.Errorf("failed to perform bulk operation in OpenSearch: %v", err)
	}

	return nil

}

func newBulkProductUpsert(p api.ProductChange) (string, error) {
	update := map[string]map[string]string{
		"update": {
			"_index": store.ProductsIndex,
			"_id":    p.Id,
		},
	}
	opData, err := json.Marshal(update)
	if err != nil {
		return "", err
	}
	doc := make(map[string]any)
	docVal := map[string]any{
		"doc":           doc,
		"doc_as_upsert": true,
	}
	doc["name"] = p.Name
	doc["description"] = p.Description
	doc["price"] = p.Price
	doc["stock"] = p.Stock
	if len(p.Pictures) > 0 {
		doc["picture"] = p.Pictures[0].Url
	} else {
		doc["picture"] = nil
	}
	docData, err := json.Marshal(docVal)
	if err != nil {
		return "", err
	}
	return string(opData) + "\n" + string(docData), nil
}
func newBulkProductDelete(p api.ProductChange) (string, error) {
	update := map[string]map[string]string{
		"delete": {
			"_index": store.ProductsIndex,
			"_id":    p.Id,
		},
	}
	opData, err := json.Marshal(update)
	if err != nil {
		return "", err
	}
	return string(opData), nil
}
