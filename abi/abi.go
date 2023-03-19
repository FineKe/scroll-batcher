package abi

import (
	_ "embed"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/pkg/errors"
	"strings"
)

var (
	//go:embed l1-abi.json
	l1abiStr string
	//go:embed l2-abi.json
	l2abiStr string

	//go:embed l1-oracle.json
	l1oracleStr string

	//go:embed l2-oracle.json
	l2oracleStr string
)

var L1Abi abi.ABI
var L2Abi abi.ABI
var L1Oracle abi.ABI
var L2Oracle abi.ABI

func init() {
	var err error

	L1Abi, err = initAbi(l1abiStr)
	if err != nil {
		panic(errors.WithMessage(err, "load abi failed"))
	}
	L2Abi, err = initAbi(l2abiStr)
	if err != nil {
		panic(errors.WithMessage(err, "load abi failed"))
	}

	L1Oracle, err = initAbi(l1oracleStr)
	if err != nil {
		panic(errors.WithMessage(err, "load abi failed"))
	}

	L2Oracle, err = initAbi(l2oracleStr)
	if err != nil {
		panic(errors.WithMessage(err, "load abi failed"))
	}
}

func initAbi(abijson string) (abi.ABI, error) {
	return abi.JSON(strings.NewReader(abijson))
}
