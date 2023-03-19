package utils

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"gotest.tools/assert"
	"testing"
)

func TestTxIsSucceed(t *testing.T) {
	client, _ := ethclient.Dial("https://goerli.infura.io/v3/edb98b2275ff4f879d083ed32e009657")
	assert.Assert(t, TxIsSucceed(client, common.HexToHash("0x4ab929a79eec75111d38705e1d18df7664a5c4c36298d381ef97bece5ca0873b")))
}
