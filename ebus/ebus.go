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
	if err := e.mq.Publisher("notifications/send"); err != nil {
		return err
	}

	return nil
}

func (e *EventBus) registerConsumers() error {
	err := e.mq.Consumer("account/create", 5, false, e.handleCreateAccount)
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

	_, err = e.ds.Account.Create(opentracing.ContextWithSpan(context.Background(), span), req.OwnerId, e.clientSecret)
	if err != nil {
		e.logger.Errorf("failed to create account: %s", err)
		return nil
	}

	return nil
}

func (e *EventBus) SendNotification(span opentracing.Span, req *notificationv1.Notification) error {
	headers := make(amqp.Table)
	ext.SpanKindRPCServer.Set(span)
	ext.Component.Set(span, "accounts")

	span.Tracer().Inject(
		span.Context(),
		opentracing.TextMap,
		mqmux.RMQHeaderCarrier(headers),
	)

	return e.mq.PublishX("notifications/send", req, headers)
}
