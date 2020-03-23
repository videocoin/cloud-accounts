package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"github.com/videocoin/cloud-accounts/service"
	"github.com/videocoin/cloud-pkg/logger"
	"github.com/videocoin/cloud-pkg/tracer"
)

var (
	ServiceName string = "accounts"
	Version     string = "dev"
)

func main() {
	err := logger.Init(ServiceName, Version)
	if err != nil {
		fmt.Printf("Failed to init logger: %s", err)
	}

	log := logrus.NewEntry(logrus.New())
	log = logrus.WithFields(logrus.Fields{
		"service": ServiceName,
		"version": Version,
	})

	closer, err := tracer.NewTracer(ServiceName)
	if err != nil {
		log.Info(err.Error())
	} else {
		defer closer.Close()
	}

	cfg := &service.Config{
		Name:    ServiceName,
		Version: Version,
	}

	err = envconfig.Process(ServiceName, cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	cfg.Logger = log

	svc, err := service.NewService(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	signals := make(chan os.Signal, 1)
	exit := make(chan bool, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signals

		log.Infof("recieved signal %s", sig)
		exit <- true
	}()

	log.Info("starting")
	go svc.Start()

	<-exit

	log.Info("stopping")
	err = svc.Stop()
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("stopped")
}
