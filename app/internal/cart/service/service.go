package service

import (
	"context"
	"errors"
	"fmt"

	oapi_codegen "github.com/bratushkadan/floral/internal/cart/presentation/generated"
	"github.com/bratushkadan/floral/internal/cart/store"
	"go.uber.org/zap"
)

type Cart struct {
	l     *zap.Logger
	store *store.Cart
}

type CartBuilder struct {
	svc Cart
}

func NewBuilder() *CartBuilder {
	return &CartBuilder{}
}

func (b *CartBuilder) Logger(l *zap.Logger) *CartBuilder {
	b.svc.l = l
	return b
}
func (b *CartBuilder) Store(store *store.Cart) *CartBuilder {
	b.svc.store = store
	return b
}

func (b *CartBuilder) Build() (*Cart, error) {
	if b.svc.store == nil {
		return nil, errors.New("store is nil")
	}

	if b.svc.l == nil {
		b.svc.l = zap.NewNop()
	}

	return &b.svc, nil
}

func (c *Cart) GetCartPositions(ctx context.Context, userId string) ([]oapi_codegen.CartGetCartPositionsResPosition, error) {
	positions, err := c.store.GetCartPositions(ctx, userId)
	if err != nil {
		return nil, err
	}
	return positions, nil
}

func (c *Cart) SetCartPosition(ctx context.Context, userId, productId string, count int) (oapi_codegen.CartSetCartPositionResPosition, error) {
	return c.store.SetCartPosition(ctx, userId, productId, count)
}
func (c *Cart) DeleteCartPosition(ctx context.Context, userId, productId string) (oapi_codegen.CartDeleteCartPositionResPosition, error) {
	return c.store.DeleteCartPosition(ctx, userId, productId)
}

func (c *Cart) ClearCart(ctx context.Context, userId string) error {
	_, err := c.store.Clear(ctx, userId)
	return err
}
func (c *Cart) ClearCarts(ctx context.Context, messages []oapi_codegen.PrivateClearCartPositionsReqMessage) error {
	return c.store.ClearMany(ctx, messages)
}

func (c *Cart) CartsPublishPositions(ctx context.Context, req oapi_codegen.PrivatePublishCartPositionsReq) error {
	positions, err := c.store.GetCartPositionsMany(ctx, req.Messages)
	if err != nil {
		return fmt.Errorf("get cart positions many: %w", err)
	}

	if err := c.store.PublishCartContents(ctx, positions); err != nil {
		return fmt.Errorf("publish cart contents: %w", err)
	}

	return nil
}
