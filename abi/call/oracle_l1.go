package call

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"scroll-batch/abi"
)

func GetL1Fee(to *common.Address, client *ethclient.Client) (*big.Int, error) {
	payload, err := abi.L1Oracle.Pack("l1BaseFee")
	if err != nil {
		return nil, err
	}

	resultBytes, err := client.CallContract(context.Background(), ethereum.CallMsg{
		To:   to,
		Data: payload,
	}, nil)

	if err != nil {
		return nil, err
	}
	n := big.NewInt(0).SetBytes(resultBytes)
	return big.NewInt(0).Mul(n, big.NewInt(16e4)), nil
}
