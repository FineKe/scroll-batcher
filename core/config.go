package core

type Config struct {
	Index           uint64 `json:"index"`
	L1ClientUrl     string `json:"l1ClientUrl"`
	L2ClientUrl     string `json:"l2ClientUrl"`
	SuperKey        string `json:"superKey"`
	Mnemonic        string `json:"mnemonic"`
	TransferAmount  int64  `json:"transferAmount"`
	DepositAmount   int64  `json:"depositAmount"`
	WithdrawAmount  int64  `json:"withdrawAmount"`
	L1oracleAddress string `json:"l1OracleAddress"`
	L2oracleAddress string `json:"l2OracleAddress"`
	DepositAddress  string `json:"depositAddress"`
	WithdrawAddress string `json:"withdrawAddress"`
}
