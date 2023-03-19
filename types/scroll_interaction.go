package types

import "github.com/ethereum/go-ethereum/common"

type ScrollInteraction struct {
	Index          uint64       `json:"index"`
	PrivateKey     string       `json:"privateKey"`
	Address        string       `json:"address"`
	DepositAmount  string       `json:"depositAmount"`
	L1Hash         *common.Hash `json:"l1Hash"`
	WithdrawAmount string       `json:"withdrawAmount"`
	L2hash         *common.Hash `json:"l2Hash"`
}

func (s ScrollInteraction) IsCompleted() bool {
	return s.L1Hash != nil && s.L2hash != nil
}
