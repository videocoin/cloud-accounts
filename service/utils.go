package service

import (
	"crypto/rand"
	"math/big"

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

func wei2Vdc(wei *big.Int) (*big.Float, error) {
	var factor, exp = big.NewInt(18), big.NewInt(10)
	exp = exp.Exp(exp, factor, nil)

	fwei := new(big.Float).SetInt(wei)

	return new(big.Float).Quo(fwei, new(big.Float).SetInt(exp)), nil
}
