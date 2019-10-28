package manager

import (
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

	return &Manager{
		eth:          opts.Eth,
		vdc:          opts.Vdc,
		bankKey:      key,
		clientSecret: opts.ClientSecret,
		tokenAddr:    opts.TokenAddr,
		ds:           opts.Ds,
		nc:           nc,
		logger:       opts.Logger,
	}, nil
}
