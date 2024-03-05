package services

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/0x726f6f6b6965/bank/internal/proto"
	"github.com/0x726f6f6b6965/bank/internal/utils"
)

var (
	bankService  *bank
	onceInitBank sync.Once

	NonceLen = 10

	ErrEmptyAccount     = errors.New("account is empty")
	ErrAccountExist     = errors.New("account already exist")
	ErrAccountNotExist  = errors.New("account not exist")
	ErrEmptyPwd         = errors.New("password is empty")
	ErrEmptyNonce       = errors.New("password is nonce")
	ErrVerify           = errors.New("account or password is not correct")
	ErrNegativeBalance  = errors.New("negative value is not allowed")
	ErrBalanceNotEnough = errors.New("balance is not enough")
	ErrFromAccount      = errors.New("from account is not correct")
	ErrToAccount        = errors.New("to account is not correct")
)

type bank struct {
	users  *userMap
	txs    *txMap
	count  uint64
	search *search
}

type userMap struct {
	sync.RWMutex
	data map[string]proto.User
}

type txMap struct {
	sync.RWMutex
	data map[uint64]proto.Transaction
}

type BankInterface interface {
	CreateAccount(ctx context.Context, user proto.User) (*proto.User, error)
	Deposit(ctx context.Context, tx proto.Transaction, nonce string) (*proto.Transaction, string, error)
	Withdraw(ctx context.Context, tx proto.Transaction, nonce string) (*proto.Transaction, string, error)
	Transaction(ctx context.Context, tx proto.Transaction, nonce string) (*proto.Transaction, string, error)
	GetNonce(ctx context.Context, account, pwd string) (string, error)
	GetBalance(ctx context.Context, account string) (int, error)
	GetTransactions(ctx context.Context, account string) ([]proto.Transaction, error)
}

func GetBankService() BankInterface {
	return bankService
}

func NewBank() BankInterface {
	onceInitBank.Do(func() {
		bankService = &bank{
			users: &userMap{
				data: make(map[string]proto.User),
			},
			txs: &txMap{
				data: make(map[uint64]proto.Transaction),
			},
			count:  1,
			search: NewSearch(),
		}
	})
	return bankService
}

func (b *bank) CreateAccount(ctx context.Context, user proto.User) (*proto.User, error) {
	if utils.IsEmpty(user.Account) {
		return nil, ErrEmptyAccount
	}

	if utils.IsEmpty(user.Password) {
		return nil, ErrEmptyPwd
	}

	if utils.IsEmpty(user.Name) {
		user.Name = "anonymous"
	}

	if user.Balance <= 0 {
		return nil, ErrNegativeBalance
	}

	nonce, err := utils.GenerateNonce(NonceLen)
	if err != nil {
		return nil, err
	}

	user.Nonce = nonce

	user.CreatedAt = time.Now().Unix()
	user.UpdatedAt = time.Now().Unix()

	b.users.Lock()
	defer b.users.Unlock()
	if _, ok := b.users.data[user.Account]; ok {
		return nil, ErrAccountExist
	}
	b.users.data[user.Account] = user
	return &user, nil
}

func (b *bank) Deposit(ctx context.Context, tx proto.Transaction, nonce string) (*proto.Transaction, string, error) {

	if !utils.IsEmpty(tx.From) {
		return nil, "", ErrFromAccount
	}

	if utils.IsEmpty(tx.To) {
		return nil, "", ErrToAccount
	}

	if tx.Amount <= 0 {
		return nil, "", ErrNegativeBalance
	}

	if utils.IsEmpty(nonce) {
		return nil, "", ErrEmptyNonce
	}

	b.users.RLock()
	user, ok := b.users.data[tx.To]
	b.users.RUnlock()
	if !ok || user.Nonce != nonce {
		return nil, "", ErrVerify
	}

	b.users.Lock()
	b.txs.Lock()
	defer b.users.Unlock()
	defer b.txs.Unlock()
	newNonce, err := utils.GenerateNonce(NonceLen)
	if err != nil {
		return nil, "", err
	}

	user.Balance += tx.Amount
	user.Nonce = newNonce
	user.UpdatedAt = time.Now().Unix()

	tx.ID = b.count
	tx.CreatedAt = time.Now().Unix()
	tx.State = proto.TransactionStateSuccess

	b.users.data[user.Account] = user
	b.txs.data[tx.ID] = tx
	atomic.AddUint64(&b.count, 1)
	b.search.Add(user.Account, tx.ID)

	return &tx, newNonce, nil
}

func (b *bank) Withdraw(ctx context.Context, tx proto.Transaction, nonce string) (*proto.Transaction, string, error) {

	if utils.IsEmpty(tx.From) {
		return nil, "", ErrFromAccount
	}

	if !utils.IsEmpty(tx.To) {
		return nil, "", ErrToAccount
	}

	if tx.Amount <= 0 {
		return nil, "", ErrNegativeBalance
	}

	if utils.IsEmpty(nonce) {
		return nil, "", ErrEmptyNonce
	}

	b.users.RLock()
	user, ok := b.users.data[tx.From]
	b.users.RUnlock()

	if !ok || user.Nonce != nonce {
		return nil, "", ErrVerify
	}

	if user.Balance < tx.Amount {
		return nil, "", ErrBalanceNotEnough
	}

	b.users.Lock()
	b.txs.Lock()
	defer b.users.Unlock()
	defer b.txs.Unlock()
	newNonce, err := utils.GenerateNonce(NonceLen)
	if err != nil {
		return nil, "", err
	}
	user.Balance -= tx.Amount
	user.Nonce = newNonce
	user.UpdatedAt = time.Now().Unix()

	tx.ID = b.count
	tx.CreatedAt = time.Now().Unix()
	tx.State = proto.TransactionStateSuccess

	b.users.data[user.Account] = user
	b.txs.data[tx.ID] = tx
	atomic.AddUint64(&b.count, 1)
	b.search.Add(user.Account, tx.ID)

	return &tx, newNonce, nil
}

func (b *bank) Transaction(ctx context.Context, tx proto.Transaction, nonce string) (*proto.Transaction, string, error) {

	if utils.IsEmpty(tx.From) {
		return nil, "", ErrFromAccount
	}

	if utils.IsEmpty(tx.To) {
		return nil, "", ErrToAccount
	}

	if tx.Amount <= 0 {
		return nil, "", ErrNegativeBalance
	}

	if utils.IsEmpty(nonce) {
		return nil, "", ErrEmptyNonce
	}

	b.users.RLock()
	fromUser, ok := b.users.data[tx.From]
	toUser, ok2 := b.users.data[tx.To]
	b.users.RUnlock()

	if !ok || fromUser.Nonce != nonce || !ok2 {
		return nil, "", ErrVerify
	}

	if fromUser.Balance < tx.Amount {
		return nil, "", ErrBalanceNotEnough
	}

	b.users.Lock()
	b.txs.Lock()
	defer b.users.Unlock()
	defer b.txs.Unlock()
	newNonce, err := utils.GenerateNonce(NonceLen)
	if err != nil {
		return nil, "", err
	}

	toUser.Balance += tx.Amount
	toUser.UpdatedAt = time.Now().Unix()

	fromUser.Balance -= tx.Amount
	fromUser.Nonce = newNonce
	fromUser.UpdatedAt = time.Now().Unix()

	tx.ID = b.count
	tx.CreatedAt = time.Now().Unix()
	tx.State = proto.TransactionStateSuccess

	b.users.data[fromUser.Account] = fromUser
	b.users.data[toUser.Account] = toUser
	b.txs.data[tx.ID] = tx
	atomic.AddUint64(&b.count, 1)
	b.search.Add(fromUser.Account, tx.ID)
	b.search.Add(toUser.Account, tx.ID)

	return &tx, newNonce, nil
}

func (b *bank) GetNonce(ctx context.Context, account string, pwd string) (string, error) {
	if utils.IsEmpty(account) {
		return "", ErrEmptyAccount
	}

	if utils.IsEmpty(pwd) {
		return "", ErrEmptyPwd
	}

	b.users.RLock()
	user, ok := b.users.data[account]
	b.users.RUnlock()

	if !ok || user.Password != pwd {
		return "", ErrVerify
	}

	b.users.Lock()
	defer b.users.Unlock()

	nonce, err := utils.GenerateNonce(NonceLen)
	if err != nil {
		return "", err
	}

	user.Nonce = nonce
	user.UpdatedAt = time.Now().Unix()

	b.users.data[account] = user

	return nonce, nil
}

func (b *bank) GetBalance(ctx context.Context, account string) (int, error) {
	if utils.IsEmpty(account) {
		return 0, ErrEmptyAccount
	}

	b.users.RLock()
	defer b.users.RUnlock()

	user, ok := b.users.data[account]
	if !ok {
		return 0, ErrAccountNotExist
	}

	return user.Balance, nil
}

func (b *bank) GetTransactions(ctx context.Context, account string) ([]proto.Transaction, error) {
	resp := []proto.Transaction{}
	if utils.IsEmpty(account) {
		return resp, ErrEmptyAccount
	}

	b.users.RLock()
	_, ok := b.users.data[account]
	b.users.RUnlock()

	if !ok {
		return resp, ErrAccountNotExist
	}

	ids := b.search.Get(account)

	b.txs.RLock()
	for _, id := range ids {
		if tx, ok := b.txs.data[id]; ok {
			resp = append(resp, tx)
		}
	}
	b.txs.RUnlock()

	return resp, nil
}
