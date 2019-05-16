package service

import (
	"github.com/sirupsen/logrus"
)

type Config struct {
	Name    string `envconfig:"-"`
	Version string `envconfig:"-"`

	RPCAddr      string `default:"0.0.0.0:5001"`
	NodeHTTPAddr string `default:"" envconfig:"NODEHTTPADDR"`
	DBURI        string `default:"root:@tcp(127.0.0.1:3306)/videocoin?charset=utf8&parseTime=True&loc=Local" envconfig:"DBURI"`
	MQURI        string `default:"amqp://guest:guest@127.0.0.1:5672" envconfig:"MQURI"`
	Secret       string `default:"secret"`

	Logger *logrus.Entry `envconfig:"-"`
}
