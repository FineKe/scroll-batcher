package utils

import (
	"context"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"math/big"
	"time"
)

func Transfer(client *ethclient.Client, from *ecdsa.PrivateKey, to common.Address, amount *big.Int) (*common.Hash, error) {
	publicKey := from.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, errors.Wrap(err, "get pending nonce failed")
	}

	gasLimit := uint64(21000) // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "get gasPrice failed")
	}

	tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, nil)
	chaiId, err := client.ChainID(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "fetch chainID failed")
	}

	tx, err = types.SignTx(tx, types.NewEIP155Signer(chaiId), from)
	if err != nil {
		return nil, errors.Wrap(err, "sign tx failed")
	}

	hash := tx.Hash()
	return &hash, client.SendTransaction(context.Background(), tx)
}

func TxIsSucceed(client *ethclient.Client, txHash common.Hash) bool {
	for i := 0; i < 20; i++ {
		time.Sleep(5 * time.Second)
		re, err := client.TransactionReceipt(context.Background(), txHash)
		if err != nil {
			continue
		}

		return re.Status == 1
	}

	return false
}
