package manager

import (
	"context"
	"sync"
	"time"

	ec "github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"github.com/videocoin/cloud-accounts/datastore"
	"github.com/videocoin/cloud-accounts/ebus"
)

type Opts struct {
	Vdc          *ec.Client
	Ds           *datastore.Datastore
	EB           *ebus.EventBus
	Logger       *logrus.Entry
	ClientSecret string
}

type Manager struct {
	vdc          *ec.Client
	ds           *datastore.Datastore
	logger       *logrus.Entry
	clientSecret string
	bTicker      *time.Ticker
	bTimeout     time.Duration
	rbLock       sync.Mutex
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
		m.rbLock.Lock()

		ctx := context.Background()
		accounts, err := m.ds.Account.List(ctx)
		if err != nil {
			m.logger.Error(err)
			time.Sleep(time.Second * 5)
			continue
		}

		for _, account := range accounts {
			_, err := m.refreshBalance(ctx, account)
			if err != nil {
				m.logger.Error(err)
				continue
			}
		}

		m.rbLock.Unlock()
	}
}
