package manager

import (
	"context"
	"time"

	ec "github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"github.com/videocoin/cloud-accounts/datastore"
	"github.com/videocoin/cloud-accounts/ebus"
	faucetcli "github.com/videocoin/cloud-pkg/faucet"
)

type Opts struct {
	Vdc          *ec.Client
	Ds           *datastore.Datastore
	EB           *ebus.EventBus
	Logger       *logrus.Entry
	ClientSecret string
	Faucet       *faucetcli.Client
}

type Manager struct {
	vdc          *ec.Client
	ds           *datastore.Datastore
	logger       *logrus.Entry
	clientSecret string
	bTicker      *time.Ticker
	bTimeout     time.Duration
	faucet       *faucetcli.Client
}

func NewManager(opts *Opts) (*Manager, error) {
	bTimeout := 10 * time.Second

	return &Manager{
		vdc:          opts.Vdc,
		clientSecret: opts.ClientSecret,
		ds:           opts.Ds,
		logger:       opts.Logger,
		bTimeout:     bTimeout,
		bTicker:      time.NewTicker(bTimeout),
		faucet:       opts.Faucet,
	}, nil
}

func (m *Manager) StartBackgroundTasks() {
	go m.startRefreshBalanceTask()
}

func (m *Manager) StopBackgroundTasks() error {
	m.bTicker.Stop()
	return nil
}

func (m *Manager) startRefreshBalanceTask() {
	for range m.bTicker.C {
		ctx := context.Background()
		accounts, err := m.ds.Account.List(ctx)
		if err != nil {
			m.logger.Errorf("failed to get accounts list: %s", err)
			continue
		}

		for _, account := range accounts {
			_, err := m.refreshBalance(ctx, account)
			if err != nil {
				m.logger.Errorf("failed to refresh balance: %s", err)
				continue
			}
		}
	}
}
