package service

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"github.com/videocoin/cloud-accounts/ebus"
	v1 "github.com/videocoin/cloud-api/notifications/v1"
	transfersv1 "github.com/videocoin/cloud-api/transfers/v1"
	usersv1 "github.com/videocoin/cloud-api/users/v1"
)

type NotificationClient struct {
	eb     *ebus.EventBus
	logger *logrus.Entry
}

func NewNotificationClient(eb *ebus.EventBus, logger *logrus.Entry) (*NotificationClient, error) {
	return &NotificationClient{
		eb:     eb,
		logger: logger,
	}, nil
}

func (c *NotificationClient) SendWithdrawSucceeded(ctx context.Context, user *usersv1.User, transfer *transfersv1.Transfer) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "SendWithdrawSucceeded")
	defer span.Finish()

	md := metautils.ExtractIncoming(ctx)

	params := map[string]string{
		"to":      user.Email,
		"address": transfer.ToAddress,
		"amount":  fmt.Sprintf("%f", transfer.Amount),
		"tx":      string(transfer.TxErc20Id),
		"domain":  md.Get("x-forwarded-host"),
	}

	notification := &v1.Notification{
		Target:   v1.NotificationTarget_EMAIL,
		Template: "user_withdraw_succeeded",
		Params:   params,
	}

	err := c.eb.SendNotification(span, notification)
	if err != nil {
		return err
	}

	return nil
}

func (c *NotificationClient) SendWithdrawFailed(ctx context.Context, user *usersv1.User, transfer *transfersv1.Transfer, reason string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "SendWithdrawFailed")
	defer span.Finish()

	md := metautils.ExtractIncoming(ctx)

	params := map[string]string{
		"to":      user.Email,
		"address": transfer.ToAddress,
		"amount":  fmt.Sprintf("%f", transfer.Amount),
		"reason":  reason,
		"domain":  md.Get("x-forwarded-host"),
	}

	notification := &v1.Notification{
		Target:   v1.NotificationTarget_EMAIL,
		Template: "user_withdraw_failed",
		Params:   params,
	}

	err := c.eb.SendNotification(span, notification)
	if err != nil {
		return err
	}

	return nil
}
