package service

import (
	"fmt"

	ec "github.com/ethereum/go-ethereum/ethclient"
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
	m   *manager.Manager
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

	vc, err := ec.Dial(cfg.RPCNodeHTTPAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial native client: %s", err.Error())
	}

	ec, err := ec.Dial(cfg.RPCEthHTTPAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial eth client: %s", err.Error())
	}

	manager, err := manager.NewManager(
		&manager.Opts{
			Ds:           ds,
			EB:           eb,
			Eth:          ec,
			Vdc:          vc,
			TokenAddr:    cfg.TokenAddr,
			BankKey:      cfg.BankKey,
			BankSecret:   cfg.BankSecret,
			ClientSecret: cfg.ClientSecret,
			Logger:       cfg.Logger.WithField("system", "manager"),
		})
	if err != nil {
		return nil, err
	}

	rpcConfig := &rpc.ServerOptions{
		Addr:         cfg.RPCAddr,
		DS:           ds,
		EB:           eb,
		Manager:      manager,
		ClientSecret: cfg.ClientSecret,
		Logger:       cfg.Logger,
	}

	rpc, err := rpc.NewServer(rpcConfig)
	if err != nil {
		return nil, err
	}

	svc := &Service{
		cfg: cfg,
		rpc: rpc,
		eb:  eb,
		m:   manager,
	}

	return svc, nil
}

func (s *Service) Start() {
	go func() {
		err := s.rpc.Start()
		if err != nil {
			s.cfg.Logger.Error(err)
		}
	}()

	go func() {
		err := s.eb.Start()
		if err != nil {
			s.cfg.Logger.Error(err)
		}
	}()

	s.m.StartBackgroundTasks()
}

func (s *Service) Stop() error {
	err := s.eb.Stop()
	if err != nil {
		return err
	}
	err = s.m.StopBackgroundTasks()
	return err
}
