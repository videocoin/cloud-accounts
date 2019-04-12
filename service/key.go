package service

import (
	"crypto/rand"

	"github.com/VideoCoin/go-videocoin/accounts/keystore"
)

type Key struct {
	Address string
	KeyFile string
}

func GenerateKey(passphrase string) (*Key, error) {
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
