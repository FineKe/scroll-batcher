package txs

import (
	"context"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"scroll-batch/abi"
)

var (
	GasLimit uint64   = 30000
	GasPrice *big.Int = big.NewInt(1000000000)
)

func SendWithdrawTx(privKey *ecdsa.PrivateKey, client *ethclient.Client, address common.Address, GasLimit uint64, GasPrice, amount *big.Int, fee *big.Int) (*types.Transaction, error) {
	pubKey, _ := privKey.Public().(*ecdsa.PublicKey)
	accountAddress := crypto.PubkeyToAddress(*pubKey)
	nonce, _ := client.PendingNonceAt(context.Background(), accountAddress)

	payload, err := abi.L2Abi.Pack("withdrawETH", amount, big.NewInt(160000))
	if err != nil {
		return nil, err
	}

	tx, err := types.SignTx(types.NewTransaction(nonce, address, big.NewInt(0).Add(fee, amount), GasLimit, GasPrice, payload), l2Signer, privKey)
	if err != nil {
		return nil, err
	}

	return tx, client.SendTransaction(context.Background(), tx)
}
