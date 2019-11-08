package datastore

import (
	"crypto/rand"
	"io"

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

func encodeToString(max int) string {
	b := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, b, max)
	if n != max {
		panic(err)
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}

var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
