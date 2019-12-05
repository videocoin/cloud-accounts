package manager

import (
	"context"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	ec "github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"github.com/videocoin/cloud-accounts/datastore"
	"github.com/videocoin/cloud-accounts/ebus"
	n "github.com/videocoin/cloud-accounts/notifications"
)

type ManagerOpts struct {
	Eth *ec.Client
	Vdc *ec.Client

	ClientSecret string
	BankKey      string
	BankSecret   string
	TokenAddr    string

	Ds     *datastore.Datastore
	EB     *ebus.EventBus
	Logger *logrus.Entry
}

type Manager struct {
	eth *ec.Client
	vdc *ec.Client

	bankKey *keystore.Key

	clientSecret string
	tokenAddr    string

	ds     *datastore.Datastore
	nc     *n.NotificationClient
	logger *logrus.Entry

	bTicker  *time.Ticker
	bTimeout time.Duration
	rbLock   sync.Mutex
}

func NewManager(opts *ManagerOpts) (*Manager, error) {
	key, err := keystore.DecryptKey([]byte(opts.BankKey), opts.BankSecret)
	if err != nil {
		return nil, err
	}

	nblogger := opts.Logger.WithField("system", "notification")
	nc, err := n.NewNotificationClient(opts.EB, nblogger)
	if err != nil {
		return nil, err
	}

	bTimeout := 60 * time.Second

	return &Manager{
		eth:          opts.Eth,
		vdc:          opts.Vdc,
		bankKey:      key,
		clientSecret: opts.ClientSecret,
		tokenAddr:    opts.TokenAddr,
		ds:           opts.Ds,
		nc:           nc,
		logger:       opts.Logger,
		bTimeout:     bTimeout,
		bTicker:      time.NewTicker(bTimeout),
	}, nil
}

func (m *Manager) StartBackgroundTasks() error {
	go m.startRefreshBalanceTask()
	return nil
}

func (m *Manager) StopBackgroundTasks() error {
	m.bTicker.Stop()
	return nil
}

func (m *Manager) startRefreshBalanceTask() error {
	for {
		select {
		case <-m.bTicker.C:
			m.rbLock.Lock()

			m.logger.Info("refresh balance")

			ctx := context.Background()
			accounts, err := m.ds.Account.List(ctx)
			if err != nil {
				m.logger.Error(err)
				time.Sleep(time.Second * 5)
				continue
			}

			for _, account := range accounts {
				logger := m.logger.WithField("id", account.Id)
				logger.Info("refreshing balance")

				_, err := m.refreshBalance(ctx, account)
				if err != nil {
					logger.WithError(err).Errorf("failed to refresh account %s balance", account.Id)
					continue
				}
			}

			m.rbLock.Unlock()
		}
	}

	return nil
}
