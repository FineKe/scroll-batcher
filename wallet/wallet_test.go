package wallet

import (
	"fmt"
	"testing"
)

func TestWallet_DeriveAccount(t *testing.T) {
	mnemonic := "jacket donkey protect feed bamboo invest embark remember train scheme subway amazing"
	w := NewWallet(mnemonic, 0)

	for i := 0; i < 1000000000; i++ {
		idx, acc := w.DeriveAccount()
		fmt.Println(idx, acc)
	}

}
