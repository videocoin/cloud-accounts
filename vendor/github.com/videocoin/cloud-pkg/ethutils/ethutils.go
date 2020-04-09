package ethutils

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

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
func WeiToEth(wei *big.Int) (*big.Float, error) {
	var factor, exp = big.NewInt(18), big.NewInt(10)
	exp = exp.Exp(exp, factor, nil)

	fwei := new(big.Float).SetInt(wei)

	return new(big.Float).Quo(fwei, new(big.Float).SetInt(exp)), nil
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
