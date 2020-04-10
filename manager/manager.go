package manager

import (
	"github.com/sirupsen/logrus"
	"github.com/videocoin/cloud-accounts/datastore"
)

type Opts struct {
	DS           *datastore.Datastore
	Logger       *logrus.Entry
	ClientSecret string
}

type Manager struct {
	ds           *datastore.Datastore
	logger       *logrus.Entry
	clientSecret string
}

func NewManager(opts *Opts) (*Manager, error) {
	return &Manager{
		logger: opts.Logger,
		ds:     opts.DS,
	}, nil
}
