package service

import (
	"github.com/sirupsen/logrus"
)

type Config struct {
	Name    string        `envconfig:"-"`
	Version string        `envconfig:"-"`
	Logger  *logrus.Entry `envconfig:"-"`

	RPCAddr      string `envconfig:"RPC_ADDR" default:"0.0.0.0:5001"`
	DBURI        string `envconfig:"DBURI" default:"root:@tcp(127.0.0.1:3306)/videocoin?charset=utf8&parseTime=True&loc=Local"`
	MQURI        string `envconfig:"MQURI" default:"amqp://guest:guest@127.0.0.1:5672"`
	ClientSecret string `default:"secret" envconfig:"CLIENT_SECRET"`
}
