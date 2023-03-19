package core

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"math/big"
	"scroll-batch/abi/call"
	"scroll-batch/abi/txs"
	"scroll-batch/types"
	"scroll-batch/utils"
	wallet2 "scroll-batch/wallet"
	"sync"
	"time"
)

var (
	BalanceNotEnoughErr = errors.New("balance not enough")
)

type scroller struct {
	l1Client                 *ethclient.Client
	l2Client                 *ethclient.Client
	superKey                 *ecdsa.PrivateKey
	wallet                   *wallet2.Wallet
	transferAmount           big.Int
	depositAmount            big.Int
	widthdrawAmount          big.Int
	scrollInterActionMap     *sync.Map
	l1oracleAddress          common.Address
	l2oracleAddress          common.Address
	depositContractAddress   common.Address
	withdrawConntractAddress common.Address
	depositCh                chan *types.ScrollInteraction
	withdrawCh               chan *types.ScrollInteraction
	ctx                      context.Context
	cancel                   context.CancelFunc
	wg                       sync.WaitGroup
	config                   Config
}

func NewScroller(config Config) *scroller {
	superAcc, err := crypto.HexToECDSA(config.SuperKey)
	if err != nil {
		panic(err)
	}

	l1Client, err := ethclient.Dial(config.L1ClientUrl)
	if err != nil {
		panic(err)
	}

	l2Client, err := ethclient.Dial(config.L2ClientUrl)
	if err != nil {
		panic(err)
	}

	scrollInterActionMap := &sync.Map{}
	bytes, err := ioutil.ReadFile("scroll_batch_interactions_output.json")
	if err == nil {
		var list []*types.ScrollInteraction
		json.Unmarshal(bytes, &list)
		for i := range list {
			item := list[i]
			scrollInterActionMap.Store(item.Index, item)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &scroller{
		l1Client:                 l1Client,
		l2Client:                 l2Client,
		superKey:                 superAcc,
		wallet:                   wallet2.NewWallet(config.Mnemonic, config.Index),
		transferAmount:           *big.NewInt(config.TransferAmount),
		depositAmount:            *big.NewInt(config.DepositAmount),
		widthdrawAmount:          *big.NewInt(config.WithdrawAmount),
		scrollInterActionMap:     scrollInterActionMap,
		l1oracleAddress:          common.HexToAddress(config.L1oracleAddress),
		l2oracleAddress:          common.HexToAddress(config.L2oracleAddress),
		depositContractAddress:   common.HexToAddress(config.DepositAddress),
		withdrawConntractAddress: common.HexToAddress(config.WithdrawAddress),
		depositCh:                make(chan *types.ScrollInteraction, 1024),
		withdrawCh:               make(chan *types.ScrollInteraction, 1024),
		ctx:                      ctx,
		cancel:                   cancel,
		config:                   config,
	}
}

func (s *scroller) StartServer() {
	s.wg.Add(4)
	go s.startTransfer()
	go s.startDeposit()
	go s.startWithdraw()
	go s.startCheckInteraction()
}

func (s *scroller) ShutDown() {
	s.cancel()
	s.wg.Wait()
	close(s.depositCh)
	close(s.withdrawCh)
	s.saveToFile()
}

func (s *scroller) saveToFile() {
	s.config.Index = s.wallet.GetCurrentIndex()
	bytes, err := json.Marshal(s.config)
	if err != nil {
		fmt.Println(err)
	} else {
		err = ioutil.WriteFile("config.json", bytes, 0666)
		if err != nil {
			log.Println("config.json 写入失败")
		}
	}

	var actions = make([]*types.ScrollInteraction, 0, 1024)

	s.scrollInterActionMap.Range(func(key, value any) bool {
		actions = append(actions, value.(*types.ScrollInteraction))
		return true
	})

	bytes, err = json.Marshal(actions)
	if err != nil {
		fmt.Println(err)
	} else {
		err = ioutil.WriteFile("scroll_batch_interactions_output.json", bytes, 0666)
		if err != nil {
			fmt.Println("scroll_batch_interactions_output.json 写入失败")
		}
	}
}

func (s *scroller) startTransfer() {
	defer s.wg.Done()

	pubKey, _ := s.superKey.Public().(*ecdsa.PublicKey)
	account := crypto.PubkeyToAddress(*pubKey)

FOR:
	for {
		select {
		case <-s.ctx.Done():
			fmt.Println("shut down transfer")
			break FOR
		default:

			time.Sleep(5 * time.Second)

			balance, err := s.l1Client.BalanceAt(context.Background(), account, nil)
			if err != nil {
				log.Println(err)
				continue
			}

			suggestPrice, err := s.l1Client.SuggestGasPrice(context.Background())
			if err != nil {
				log.Println(err)
				continue
			}

			fee := big.NewInt(0).Mul(suggestPrice, big.NewInt(2100))
			amount := fee.Add(fee, &s.depositAmount)

			// balance not enough
			if balance.Cmp(amount) == -1 {
				continue
			}

			// drive a new account
			idx, account := s.wallet.DeriveAccount()
			action := types.ScrollInteraction{
				Index:          idx,
				PrivateKey:     s.wallet.GetAccountPriKey(account),
				Address:        account.Address.String(),
				DepositAmount:  "",
				L1Hash:         nil,
				WithdrawAmount: "",
				L2hash:         nil,
			}

			hash, err := utils.Transfer(s.l1Client, s.superKey, account.Address, &s.transferAmount)
			if err != nil {
				continue
			}

			s.scrollInterActionMap.Store(idx, &action)
			log.Println(fmt.Sprintf("transfer %s wei to %s, tx: %s\n", s.transferAmount.String(), account.Address, hash))
		}

	}
}

func (s *scroller) startDeposit() {
	defer s.wg.Done()

FOR:
	for {
		select {
		case <-s.ctx.Done():
			fmt.Println("shut down StartDeposit...")
			break FOR
		case action := <-s.depositCh:
			if action.L1Hash != nil {
				continue
			}

			privateKey, _ := crypto.HexToECDSA(action.PrivateKey)
			tx, err := sendDeposit(privateKey, s.l1Client, s.l2oracleAddress, s.depositContractAddress, s.depositAmount)
			if err != nil {
				if err == BalanceNotEnoughErr {
					continue
				} else {
					log.Println(err)
				}
			} else {
				action.L1Hash = tx
				action.DepositAmount = s.depositAmount.String()
				s.scrollInterActionMap.Store(action.Index, action)
				log.Println("send deposit tx:", tx)
			}
		}

		time.Sleep(5 * time.Second)
	}
}

func (s *scroller) startWithdraw() {
	defer s.wg.Done()

FOR:
	for {
		select {
		case <-s.ctx.Done():
			fmt.Println("shut down StartWithdraw...")
			break FOR
		case action := <-s.withdrawCh:
			privateKey, _ := crypto.HexToECDSA(action.PrivateKey)
			if action.L1Hash == nil || action.L2hash != nil {
				continue
			}
			tx, err := sendWithdraw(privateKey, s.l2Client, s.l1oracleAddress, s.withdrawConntractAddress, s.widthdrawAmount)
			if err != nil {
				if err == BalanceNotEnoughErr {
					continue
				} else {
					log.Println(err)
				}
			} else {
				action.L2hash = tx
				action.WithdrawAmount = s.widthdrawAmount.String()
				s.scrollInterActionMap.Store(action.Index, action)
				log.Println("send withdraw tx: ", tx)
			}

		}

		time.Sleep(5 * time.Second)
	}
}

func (s *scroller) startCheckInteraction() {
	defer s.wg.Done()

FOR:
	for {
		select {
		case <-s.ctx.Done():
			fmt.Println("shut down check interation")
			break FOR
		default:
			if len(s.depositCh) == 0 {
				s.scrollInterActionMap.Range(func(key, value any) bool {
					v := value.(*types.ScrollInteraction)
					if v.L1Hash == nil {
						s.depositCh <- v
					}
					return true
				})
			}

			if len(s.withdrawCh) == 0 {
				s.scrollInterActionMap.Range(func(key, value any) bool {
					v := value.(*types.ScrollInteraction)
					if v.L1Hash != nil && v.L2hash == nil {
						s.withdrawCh <- v
					}
					return true
				})
			}

			time.Sleep(60 * time.Second)
		}
	}
}

func sendWithdraw(privKey *ecdsa.PrivateKey, client *ethclient.Client, l1oracleAddress, withdrawConntractAddress common.Address, amount big.Int) (*common.Hash, error) {
	suggestGP, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		suggestGP = txs.GasPrice
	}
	l1fee, err := call.GetL1Fee(&l1oracleAddress, client)
	if err != nil {
		return nil, err
	}

	gasLimit := uint64(300000)

	sum := big.NewInt(0).Mul(big.NewInt(300000), suggestGP)
	sum = sum.Add(sum, l1fee)
	sum = sum.Add(sum, &amount)
	pubKey, _ := privKey.Public().(*ecdsa.PublicKey)
	accountAddress := crypto.PubkeyToAddress(*pubKey)
	balance, err := client.BalanceAt(context.Background(), accountAddress, nil)
	if err != nil {
		return nil, BalanceNotEnoughErr
	}
	if balance.Cmp(sum) == -1 {
		return nil, BalanceNotEnoughErr
	}

	tx, err := txs.SendWithdrawTx(privKey, client, withdrawConntractAddress, gasLimit, suggestGP, &amount, l1fee)
	if err != nil {
		return nil, err
	}

	hash := tx.Hash()
	return &hash, err
}

func sendDeposit(privKey *ecdsa.PrivateKey, l1Client *ethclient.Client, l2oracleAddress, depositContractAddress common.Address, amount big.Int) (*common.Hash, error) {
	suggestGP, err := l1Client.SuggestGasPrice(context.Background())
	if err != nil {
		suggestGP = txs.GasPrice
	}

	l2fee, err := call.GetL2Fee(&l2oracleAddress, l1Client)
	if err != nil {
		return nil, err
	}
	gasLimit := uint64(300000)

	sum := big.NewInt(0).Mul(big.NewInt(300000), suggestGP)
	sum = sum.Add(sum, l2fee)
	sum = sum.Add(sum, &amount)
	pubKey, _ := privKey.Public().(*ecdsa.PublicKey)
	accountAddress := crypto.PubkeyToAddress(*pubKey)
	balance, err := l1Client.BalanceAt(context.Background(), accountAddress, nil)
	if err != nil {
		return nil, BalanceNotEnoughErr
	}
	if balance.Cmp(sum) == -1 {
		return nil, BalanceNotEnoughErr
	}

	tx, err := txs.SendDepositTx(privKey, l1Client, depositContractAddress, gasLimit, suggestGP, &amount, l2fee)
	if err != nil {
		return nil, err
	}

	hash := tx.Hash()

	return &hash, nil
}
