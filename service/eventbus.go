package service

import (
	"encoding/json"

	v1 "github.com/VideoCoin/cloud-api/accounts/v1"
	"github.com/VideoCoin/cloud-pkg/mqmux"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type EventBus struct {
	logger *logrus.Entry
	mq     *mqmux.WorkerMux
	ds     *Datastore
	secret string
}

func NewEventBus(mq *mqmux.WorkerMux, ds *Datastore, secret string, logger *logrus.Entry) (*EventBus, error) {
	return &EventBus{
		logger: logger,
		mq:     mq,
		ds:     ds,
		secret: secret,
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

	if req.OwnerId == "" {
		e.logger.Error("failed to create account: owner is empty")
		return nil
	}

	_, err = e.ds.Account.Create(req.OwnerId, e.secret)
	if err != nil {
		e.logger.Errorf("failed to create account: %s", err)
		return nil
	}

	return nil
}
