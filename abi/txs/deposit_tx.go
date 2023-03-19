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

var l1Signer = types.NewEIP155Signer(big.NewInt(5))
var l2Signer = types.NewEIP155Signer(big.NewInt(534353))

func SendDepositTx(privKey *ecdsa.PrivateKey, client *ethclient.Client, address common.Address, GasLimit uint64, GasPrice, amount *big.Int, fee *big.Int) (*types.Transaction, error) {

	payload, err := abi.L1Abi.Pack("depositETH", amount, big.NewInt(40000))
	if err != nil {
		return nil, err
	}
	pubKey, ok := privKey.Public().(*ecdsa.PublicKey)
	if !ok {
		panic(err)
	}
	accountAddress := crypto.PubkeyToAddress(*pubKey)
	nonce, _ := client.PendingNonceAt(context.Background(), accountAddress)

	tx, err := types.SignTx(types.NewTransaction(nonce, address, big.NewInt(0).Add(fee, amount), GasLimit, GasPrice, payload), l1Signer, privKey)
	if err != nil {
		return nil, err
	}

	return tx, client.SendTransaction(context.Background(), tx)
}
