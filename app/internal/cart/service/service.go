package service

import (
	"context"
	"errors"

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

func (c *Cart) GetCartPositions(ctx context.Context, userId string) (oapi_codegen.CartGetCartPositionsRes, error) {
	positions, err := c.store.GetCartPositions(ctx, userId)
	if err != nil {
		return oapi_codegen.CartGetCartPositionsRes{}, err
	}
	return oapi_codegen.CartGetCartPositionsRes{Positions: positions}, nil
}

func (c *Cart) SetCartPosition(ctx context.Context, userId, productId string, count int) {
	return c.store.SetCartPosition(ctx, userId, productId, count)
}
func (c *Cart) DeleteCartPosition(ctx context.Context) {}

func (c *Cart) ClearCart(ctx context.Context)  {}
func (c *Cart) ClearCarts(ctx context.Context) {}

func (c *Cart) CartsPublishPositions(ctx context.Context, req oapi_codegen.PrivateCartPublishContentsJSONRequestBody) error {
	// positions, err := c.store.GetCartPositionsMany(ctx)

	// ydbtopic.Produce(ctx context.Context, w *topicwriter.Writer, msgs ...[]byte)
	return nil
}
