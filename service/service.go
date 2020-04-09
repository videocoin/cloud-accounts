package service

import (
	"fmt"

	ec "github.com/ethereum/go-ethereum/ethclient"
	"github.com/videocoin/cloud-accounts/datastore"
	"github.com/videocoin/cloud-accounts/ebus"
	"github.com/videocoin/cloud-accounts/manager"
	"github.com/videocoin/cloud-accounts/rpc"
	"github.com/videocoin/cloud-pkg/mqmux"
	faucetcli "github.com/videocoin/go-faucet/client"
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

	faucet := faucetcli.NewClient(cfg.FaucetURL)

	manager, err := manager.NewManager(
		&manager.Opts{
			Ds:           ds,
			EB:           eb,
			Vdc:          vc,
			ClientSecret: cfg.ClientSecret,
			Logger:       cfg.Logger.WithField("system", "manager"),
			Faucet:       faucet,
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

func (s *Service) Start(errCh chan error) {
	go func() {
		errCh <- s.rpc.Start()
	}()

	go func() {
		errCh <- s.eb.Start()
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
