package ethutils

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

// AddressHex creates addressHex
func AddressHex(i int) string {
	addressHex, _ := PrivateKeyHexToAddressHex(PrivateKeyHex(i))
	return addressHex
}

// PrivateKeyHex creates privateKeyHex
func PrivateKeyHex(i int) string {
	return fmt.Sprintf("%064d", i)
}

// PrivateKeyHexToAddressHex converts from privatekey to address
func PrivateKeyHexToAddressHex(privateKeyHex string) (string, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", err
	}
	addressHex := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	return addressHex, nil
}

const ethToWei = float64(1000000000000000000)

// EthToWei converts from ETH to Wei
func EthToWei(ether float64) *big.Int {
	if ether < 1 {
		return big.NewInt(int64(ether * ethToWei))
	}
	var result big.Int
	result.Mul(big.NewInt(int64(ether)), big.NewInt(int64(ethToWei)))
	return &result
}

// WeiToEth converts from Wei to ETH
func WeiToEth(wei *big.Int) float64 {
	eth := big.NewInt(1000000000000000000)
	if wei.Cmp(eth) == 1 {
		var result big.Int
		result.Div(wei, eth)
		return float64(result.Int64())
	}
	return float64(wei.Uint64()) / ethToWei
}

// ParseInt64 parse hex string value to int64
func ParseInt64(value string) (int64, error) {
	i, err := strconv.ParseInt(strings.TrimPrefix(value, "0x"), 16, 64)
	if err != nil {
		return 0, err
	}

	return i, nil
}

// ParseUint64 parse hex string value to int64
func ParseUint64(value string) (uint64, error) {
	i, err := strconv.ParseUint(strings.TrimPrefix(value, "0x"), 16, 64)
	if err != nil {
		return 0, err
	}

	return i, nil
}

// ParseBigInt parse hex string value to big.Int
func ParseBigInt(value string) (big.Int, error) {
	i := big.Int{}
	_, err := fmt.Sscan(value, &i)

	return i, err
}

// IntToHex convert int to hexadecimal representation
func IntToHex(i int) string {
	return fmt.Sprintf("0x%x", i)
}

// BigToHex covert big.Int to hexadecimal representation
func BigToHex(bigInt big.Int) string {
	return "0x" + strings.TrimPrefix(fmt.Sprintf("%x", bigInt.Bytes()), "0")
}
