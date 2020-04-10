package service

import (
	"github.com/videocoin/cloud-accounts/datastore"
	"github.com/videocoin/cloud-accounts/ebus"
	"github.com/videocoin/cloud-accounts/manager"
	"github.com/videocoin/cloud-accounts/rpc"
	"github.com/videocoin/cloud-pkg/mqmux"
)

type Service struct {
	cfg *Config
	rpc *rpc.Server
	eb  *ebus.EventBus
}

func NewService(cfg *Config) (*Service, error) {
	ds, err := datastore.NewDatastore(cfg.DBURI)
	if err != nil {
		return nil, err
	}

	mq, err := mqmux.NewWorkerMux(cfg.MQURI, cfg.Name)
	if err != nil {
		return nil, err
	}
	mq.Logger = cfg.Logger.WithField("system", "mq")

	eblogger := cfg.Logger.WithField("system", "eventbus")
	eb, err := ebus.NewEventBus(mq, ds, cfg.ClientSecret, eblogger)
	if err != nil {
		return nil, err
	}

	manager, err := manager.NewManager(
		&manager.Opts{
			Logger:       cfg.Logger.WithField("system", "manager"),
			ClientSecret: cfg.ClientSecret,
			DS:           ds,
		})
	if err != nil {
		return nil, err
	}

	rpcConfig := &rpc.ServerOptions{
		Logger:  cfg.Logger,
		Addr:    cfg.RPCAddr,
		DS:      ds,
		Manager: manager,
	}

	rpc, err := rpc.NewServer(rpcConfig)
	if err != nil {
		return nil, err
	}

	svc := &Service{
		cfg: cfg,
		rpc: rpc,
		eb:  eb,
	}

	return svc, nil
}

func (s *Service) Start(errCh chan error) {
	go func() {
		errCh <- s.rpc.Start()
	}()

	go func() {
		errCh <- s.eb.Start()
	}()
}

func (s *Service) Stop() error {
	return s.eb.Stop()
}
