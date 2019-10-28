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
	rpc *rpc.RpcServer
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

	vc, err := ec.Dial(cfg.RPCNodeHTTPAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial native client: %s", err.Error())
	}

	ec, err := ec.Dial(cfg.RPCEthHTTPAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial eth client: %s", err.Error())
	}

	manager, err := manager.NewManager(
		&manager.ManagerOpts{
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

	rpcConfig := &rpc.RpcServerOptions{
		Addr:         cfg.RPCAddr,
		DS:           ds,
		EB:           eb,
		Manager:      manager,
		ClientSecret: cfg.ClientSecret,
		Logger:       cfg.Logger,
	}

	rpc, err := rpc.NewRpcServer(rpcConfig)
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

func (s *Service) Start() error {
	go s.rpc.Start()
	go s.eb.Start()
	return nil
}

func (s *Service) Stop() error {
	s.eb.Stop()
	return nil
}
