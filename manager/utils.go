package manager

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/videocoin/cloud-pkg/token"
)

var (
	ErrUserNativeBalanceInsufficient    = errors.New("user native balance insufficient")
	ErrUserNativeBalanceGasInsufficient = errors.New("user native balance gas insufficient")
	ErrBankErcBalanceInsufficient       = errors.New("bank erc balance insufficient")
	ErrBankErcBalanceGasInsufficient    = errors.New("bank erc balance gas insufficient")
)

func checkUserBalance(client *ethclient.Client, address common.Address, transferAmount *big.Int) error {
	userNativeBalance, err := client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		return fmt.Errorf("failed to get user native balance: %s", err.Error())
	}

	// check the user has enough native tokens to perform the refund
	if userNativeBalance.Cmp(transferAmount) < 0 {
		return ErrUserNativeBalanceInsufficient
	}

	// check the user has enough gas to perform the refund
	gasCostEstimate := big.NewInt(100000000000000000)
	minBalance := gasCostEstimate.Add(gasCostEstimate, transferAmount)
	if userNativeBalance.Cmp(minBalance) < 0 {
		return ErrUserNativeBalanceGasInsufficient
	}

	return nil
}

func checkBankBalance(client *ethclient.Client, address, tokenAddr common.Address, transferAmount *big.Int) error {
	tokenInstance, err := token.NewToken(tokenAddr, client)
	if err != nil {
		return fmt.Errorf("failed to create token instance: %s", err.Error())
	}

	bankTokenBalance, err := tokenInstance.BalanceOf(&bind.CallOpts{}, address)
	if err != nil {
		return fmt.Errorf("failed to get bank token balance: %s", err.Error())
	}

	fmt.Printf("bank %s token balance is %d\n", address, bankTokenBalance.Int64())

	// check the bank has enough erc tokens to perform the refund; if not, raise alert and fund our bank account
	if bankTokenBalance.Cmp(transferAmount) < 0 {
		return ErrBankErcBalanceInsufficient
	}

	bankEthBalance, err := client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		return fmt.Errorf("failed to get bank eth balance: %s", err.Error())
	}

	// check that the bank actually has enough eth to cover gas cost for withdrawal tx
	gasCostEstimate := big.NewInt(100000000000000000)
	if bankEthBalance.Cmp(gasCostEstimate) < 0 {
		return ErrBankErcBalanceGasInsufficient
	}

	return nil
}

func waitMinedAndCheck(client *ethclient.Client, tx *types.Transaction) error {
	cancelCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	receipt, err := bind.WaitMined(cancelCtx, client, tx)
	if err != nil {
		return err
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("transaction %s failed", tx.Hash().String())
	}

	return nil
}

func execNativeTransaction(client *ethclient.Client, key *keystore.Key, toAddress common.Address, amount *big.Int) (*types.Transaction, error) {
	nonce, err := client.PendingNonceAt(context.Background(), key.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %s", err.Error())
	}

	tx := types.NewTransaction(nonce, toAddress, amount, uint64(21000), big.NewInt(10000000000), nil)

	// chainID, err := client.ChainID(context.Background())
	// if err != nil {
	// 	return fmt.Errorf("failed to create chain id: %s", err.Error())
	// }

	// signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), key.PrivateKey)
	signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, key.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction %s: %s", tx.Hash().Hex(), err.Error())
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction %s: %s", tx.Hash().Hex(), err.Error())
	}

	return signedTx, nil
}

func execErc20Transaction(client *ethclient.Client, key *keystore.Key, toAddress, tokenAddr common.Address, amount *big.Int) (*types.Transaction, error) {
	tokenInstance, err := token.NewToken(tokenAddr, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create token instance: %s", err.Error())
	}

	auth := bind.NewKeyedTransactor(key.PrivateKey)
	auth.Value = big.NewInt(0)
	auth.GasPrice = big.NewInt(2000000000)

	// transfer erc tokens to user on ethereum net
	tx, err := tokenInstance.Transfer(auth, toAddress, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to transfer tokens %s: %s", tx.Hash().Hex(), err.Error())
	}

	return tx, nil
}
