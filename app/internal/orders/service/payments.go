package service

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	oapi_codegen "github.com/bratushkadan/floral/internal/orders/presentation/generated"
)

var (
	ErrYoomoneyPaymentNotificationValidation           = errors.New("invalid payment notificaiton input parameter")
	ErrYoomoneyPaymentNotificationIntegrityCheckFailed = errors.New("payment notification integrity check failed")
)

type ProcessYoomoneyPaymentNotificationReq struct {
	NotificationType string
	OperationId      string
	Amount           string
	Currency         string
	Datetime         string
	Sender           string
	Codepro          string
	Label            string

	Sha1Hash string
}

func (s *Orders) ProcessYoomoneyPaymentNotification(ctx context.Context, req ProcessYoomoneyPaymentNotificationReq) error {
	amount, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil {
		return fmt.Errorf(
			`%w: invalid "amount" field value`,
			ErrYoomoneyPaymentNotificationValidation,
		)
	}

	paymentTime, err := time.Parse(time.RFC3339, req.Datetime)
	if err != nil {
		return fmt.Errorf(
			`%w: invalid "datetime" field value`,
			ErrYoomoneyPaymentNotificationValidation,
		)
	}

	currency, err := strconv.Atoi(req.Currency)
	if err != nil {
		return fmt.Errorf(
			`%w: invalid "currency" field value`,
			ErrYoomoneyPaymentNotificationValidation,
		)
	}

	integrityCheckString := strings.Join([]string{
		req.NotificationType,
		req.OperationId,
		req.Amount,
		req.Currency,
		req.Datetime,
		req.Sender,
		req.Codepro,
		s.yoomoneyPaymentNotificationSecret,
		req.Label,
	}, "&")
	integrityCheckSha1 := sha1.New()
	integrityCheckSha1.Write([]byte(integrityCheckString))
	hashBytes := integrityCheckSha1.Sum(nil)
	integrityHashString := hex.EncodeToString(hashBytes)

	if integrityHashString != req.Sha1Hash {
		return ErrYoomoneyPaymentNotificationIntegrityCheckFailed
	}

	labelParts := strings.Split(req.Label, ":")
	orderId := labelParts[0]
	// FIXME:
	orderId = "5f81f6fe-0b02-430d-8a95-38d8d3f55759"
	if orderId == "" {
		return fmt.Errorf(
			`%w: invalid empty "label" field`,
			ErrYoomoneyPaymentNotificationValidation,
		)
	}

	order, err := s.store.GetOrder(ctx, orderId)
	if err != nil {
		return fmt.Errorf("find order: %v", err)
	}
	if order == nil {
		return fmt.Errorf(
			`%w: no order found with order_id="%s"`,
			ErrYoomoneyPaymentNotificationValidation,
			orderId,
		)
	}

	if err := s.store.ProduceProcessedPaymentsNotificationsMessages(ctx, oapi_codegen.PrivateOrderProcessPaymentReqMessage{
		OrderId:         orderId,
		CurrencyIso4217: currency,
		Datetime:        paymentTime,
		Amount:          amount,
		ProviderMeta: map[string]any{
			"yoomoney": map[string]any{
				"notification_type": req.NotificationType,
				"operation_id":      req.OperationId,
				"amount":            req.Amount,
				"currency":          req.Currency,
				"datetime":          req.Datetime,
				"sender":            req.Sender,
				"codepro":           req.Codepro,
				"label":             req.Label,
			},
		},
	}); err != nil {
		return fmt.Errorf("produce processed payment notification message: %v", err)
	}

	return nil
}
