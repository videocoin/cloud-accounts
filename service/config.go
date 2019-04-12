package service

import (
	"github.com/sirupsen/logrus"
)

type Config struct {
	Name    string `envconfig:"-"`
	Version string `envconfig:"-"`

	RPCAddr string `default:"0.0.0.0:5001"`
	DBURI   string `default:"mysql:mysql@/vc-user?charset=utf8&parseTime=True&loc=Local"`
	MQURI   string `default:"amqp://rabbitmq:bitnami@127.0.0.1:5672"`
	Secret  string `default:"secret"`

	Logger *logrus.Entry `envconfig:"-"`
}
