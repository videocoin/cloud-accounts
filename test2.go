package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/videocoin/cloud-pkg/bcops"
)

func decryptAndAuth(client *ethclient.Client, keyjson, secret string) (*keystore.Key, *bind.TransactOpts, error) {
	key, err := keystore.DecryptKey([]byte(keyjson), secret)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt key: %s", err.Error())
	}

	auth, err := bcops.GetBCAuth(client, key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to auth key: %s", err.Error())
	}

	return key, auth, nil
}

func main() {
	key := "{\"address\":\"002c938fc7461db28cd13ab1f595b5c31f81fd26\",\"crypto\":{\"cipher\":\"aes-128-ctr\",\"ciphertext\":\"6172d91dfe8f1414135d14091c3f009cebca518d3b3436ce56fe5de743fdb057\",\"cipherparams\":{\"iv\":\"b3e021bac13cefb1e1684c8f47e17ca6\"},\"kdf\":\"scrypt\",\"kdfparams\":{\"dklen\":32,\"n\":262144,\"p\":1,\"r\":8,\"salt\":\"5d2ca48d75b873b64e7375961a53641ffd38ca4e4045a1a78286732249805b78\"},\"mac\":\"7a3eb545ba6ed0080e16d8125cc98766df2ec16be7dc997a3eb9ff4fd89a700c\"},\"id\":\"03fb4250-2b12-4580-886c-a61ee863edff\",\"version\":3}"
	secret := "secret"

	bKey := "{\"address\":\"b51197bfaa0fcdda4c98488f9959a13d45665309\",\"crypto\":{\"cipher\":\"aes-128-ctr\",\"ciphertext\":\"e91042ffad9849f80963c3506935afb546d62dec769dceaece96db3569cfb910\",\"cipherparams\":{\"iv\":\"1d85836b6600ce0b309268bb60394be4\"},\"kdf\":\"scrypt\",\"kdfparams\":{\"dklen\":32,\"n\":262144,\"p\":1,\"r\":8,\"salt\":\"4824a15bab4de8542bea6307ba7551de04e8c108a2f6a3e318e68dab7ba4cdea\"},\"mac\":\"69b0e9fc5d4eda78c3b388bbb18585f2541526d62a2a7700602e8a8b13587068\"},\"id\":\"2b60939d-fbfc-445b-bfb1-55e06efa8a10\",\"version\":3}"
	bSecret := "e6v4x7axa9"

	//client, err := ethclient.Dial("https://rinkeby.infura.io/v3/2133def9c46e42269dc76cff5338643a")
	client, err := ethclient.Dial("http://admin:VideoCoinS3cr3t@rpc.dev.videocoin.network")
	if err != nil {
		log.Fatal(err)
	}

	_, bankAuth, err := decryptAndAuth(client, bKey, bSecret)
	if err != nil {
		log.Fatal(err)
	}

	userKey, userAuth, err := decryptAndAuth(client, key, secret)
	if err != nil {
		log.Fatal(err)
	}

	nonce, err := client.PendingNonceAt(context.Background(), userAuth.From)
	if err != nil {
		log.Fatal(err)
	}

	amount := big.NewInt(1000000000000000000)
	tx := types.NewTransaction(nonce, bankAuth.From, amount, uint64(21000), big.NewInt(10000000000), nil)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("chain id is %d\n", chainID.Int64())

	signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, userKey.PrivateKey)
	if err != nil {
		log.Fatal(err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}
}
