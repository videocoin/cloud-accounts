package service

import (
	"github.com/sirupsen/logrus"
)

type Config struct {
	Name    string `envconfig:"-"`
	Version string `envconfig:"-"`

	RPCAddr         string `default:"0.0.0.0:5001" envconfig:"RPC_ADDR"`
	RPCNodeHTTPAddr string `default:"" envconfig:"RPC_NODE_HTTP_ADDR"`
	RPCEthHTTPAddr  string `default:"" envconfig:"RPC_ETH_HTTP_ADDR"`
	DBURI           string `default:"root:@tcp(127.0.0.1:3306)/videocoin?charset=utf8&parseTime=True&loc=Local" envconfig:"DBURI"`
	MQURI           string `default:"amqp://guest:guest@127.0.0.1:5672" envconfig:"MQURI"`

	ClientSecret string `default:"secret" envconfig:"CLIENT_SECRET"`
	BankKey      string `default:"" envconfig:"BANK_KEY"`
	BankSecret   string `default:"" envconfig:"BANK_SECRET"`

	TokenAddr string `default:"" envconfig:"TOKEN_ADDR"`

	Logger *logrus.Entry `envconfig:"-"`
}
