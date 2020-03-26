package datastore

import (
	"crypto/rand"

	"github.com/ethereum/go-ethereum/accounts/keystore"
)

type Key struct {
	Address string
	KeyFile string
}

func generateKey(passphrase string) (*Key, error) {
	key := keystore.NewKeyForDirectICAP(rand.Reader)

	json, err := keystore.EncryptKey(key, passphrase, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return nil, err
	}

	return &Key{
		Address: key.Address.String(),
		KeyFile: string(json),
	}, nil
}
