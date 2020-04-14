package ebus

import (
	"context"
	"encoding/json"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/videocoin/cloud-accounts/datastore"
	v1 "github.com/videocoin/cloud-api/accounts/v1"
	notificationv1 "github.com/videocoin/cloud-api/notifications/v1"
	"github.com/videocoin/cloud-pkg/mqmux"
)

type EventBus struct {
	logger       *logrus.Entry
	mq           *mqmux.WorkerMux
	ds           *datastore.Datastore
	clientSecret string
}

func NewEventBus(mq *mqmux.WorkerMux, ds *datastore.Datastore, clientSecret string, logger *logrus.Entry) (*EventBus, error) {
	return &EventBus{
		logger:       logger,
		mq:           mq,
		ds:           ds,
		clientSecret: clientSecret,
	}, nil
}

func (e *EventBus) Start() error {
	err := e.registerPublishers()
	if err != nil {
		return err
	}

	err = e.registerConsumers()
	if err != nil {
		return err
	}

	return e.mq.Run()
}

func (e *EventBus) Stop() error {
	return e.mq.Close()
}

func (e *EventBus) registerPublishers() error {
	if err := e.mq.Publisher("notifications.send"); err != nil {
		return err
	}

	if err := e.mq.Publisher("accounts.events"); err != nil {
		return err
	}

	return nil
}

func (e *EventBus) registerConsumers() error {
	err := e.mq.Consumer("accounts.create", 5, false, e.handleCreateAccount)
	if err != nil {
		return err
	}

	return nil
}

func (e *EventBus) handleCreateAccount(d amqp.Delivery) error {
	req := new(v1.AccountRequest)
	err := json.Unmarshal(d.Body, req)
	if err != nil {
		return err
	}

	var span opentracing.Span
	tracer := opentracing.GlobalTracer()
	spanCtx, err := tracer.Extract(opentracing.TextMap, mqmux.RMQHeaderCarrier(d.Headers))

	if err != nil {
		span = tracer.StartSpan("handleCreateAccount")
	} else {
		span = tracer.StartSpan("handleCreateAccount", ext.RPCServerOption(spanCtx))
	}

	defer span.Finish()

	span.SetTag("owner_id", req.OwnerId)

	if req.OwnerId == "" {
		e.logger.Error("failed to create account: owner is empty")
		return nil
	}

	ctx := opentracing.ContextWithSpan(context.Background(), span)
	account, err := e.ds.Account.Create(ctx, req.OwnerId, e.clientSecret)
	if err != nil {
		e.logger.Errorf("failed to create account: %s", err)
		return nil
	}

	err = e.EmitAccountCreated(ctx, account)
	if err != nil {
		e.logger.Errorf("failed to emit account created: %s", err)
		return nil
	}

	return nil
}

func (e *EventBus) SendNotification(span opentracing.Span, req *notificationv1.Notification) error {
	headers := make(amqp.Table)
	ext.SpanKindRPCServer.Set(span)
	ext.Component.Set(span, "accounts")

	err := span.Tracer().Inject(
		span.Context(),
		opentracing.TextMap,
		mqmux.RMQHeaderCarrier(headers),
	)
	if err != nil {
		return err
	}

	return e.mq.PublishX("notifications.send", req, headers)
}

func (e *EventBus) EmitAccountCreated(ctx context.Context, account *datastore.Account) error {
	headers := make(amqp.Table)

	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		ext.SpanKindRPCServer.Set(span)
		ext.Component.Set(span, "transcoder")
		err := span.Tracer().Inject(
			span.Context(),
			opentracing.TextMap,
			mqmux.RMQHeaderCarrier(headers),
		)
		if err != nil {
			e.logger.Errorf("failed to span inject: %s", err)
		}
	}
	event := &v1.Event{
		Type:    v1.EventTypeAccountCreated,
		UserID:  account.UserID,
		Address: account.Address,
	}

	err := e.mq.PublishX("accounts.events", event, headers)
	if err != nil {
		return err
	}

	return nil
}
