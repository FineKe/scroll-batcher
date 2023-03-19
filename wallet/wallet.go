package wallet

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"sync"
)

func NewWallet(mnemonic string, startIndex uint64) *Wallet {
	wallet, _ := hdwallet.NewFromMnemonic(mnemonic)

	w := &Wallet{
		mnemonic:     mnemonic,
		wallet:       wallet,
		currentIndex: startIndex,
		mux:          sync.Mutex{},
	}

	return w
}

type Wallet struct {
	mnemonic     string
	currentIndex uint64
	wallet       *hdwallet.Wallet
	mux          sync.Mutex
}

func (w *Wallet) DeriveAccount() (uint64, accounts.Account) {
	w.mux.Lock()
	defer func() {
		w.currentIndex++
		w.mux.Unlock()
	}()

	hdpath := hdwallet.MustParseDerivationPath(fmt.Sprintf("m/44'/60'/0'/0/%d", w.currentIndex))
	account, _ := w.wallet.Derive(hdpath, true)
	return w.currentIndex, account
}

func (w *Wallet) GetAccountPriKey(account accounts.Account) string {
	hex, _ := w.wallet.PrivateKeyHex(account)
	return hex
}

func (w *Wallet) GetCurrentIndex() uint64 {
	return w.currentIndex
}
