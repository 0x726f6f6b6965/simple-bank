package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/0x726f6f6b6965/bank/internal/proto"
	"github.com/google/uuid"
)

var ctx context.Context

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	ctx = context.Background()
	fmt.Printf("\033[1;33m%s\033[0m", "> Setup completed\n")
}

func teardown() {
	fmt.Printf("\033[1;33m%s\033[0m", "> Teardown completed")
	fmt.Printf("\n")
}
func TestGetNonce(t *testing.T) {
	service := &bank{
		users: &userMap{
			data: make(map[string]proto.User),
		},
		txs: &txMap{
			data: make(map[uint64]proto.Transaction),
		},
	}

	_, err := service.GetNonce(ctx, "", "")
	if !errors.Is(err, ErrEmptyAccount) {
		t.Fatalf("Expected error: %v, got: %v", ErrEmptyAccount, err)
	}

	service.users.Lock()
	service.users.data["test"] = proto.User{
		Account:  "test",
		Balance:  100,
		Password: "test-pwd",
		Name:     "test-user",
	}
	service.users.Unlock()

	_, err = service.GetNonce(ctx, "test", "")
	if !errors.Is(err, ErrEmptyPwd) {
		t.Fatalf("Expected error: %v, got: %v", ErrEmptyPwd, err)
	}

	_, err = service.GetNonce(ctx, "test", "t")
	if !errors.Is(err, ErrVerify) {
		t.Fatalf("Expected error: %v, got: %v", ErrVerify, err)
	}

	nonce, err := service.GetNonce(ctx, "test", "test-pwd")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	service.users.RLock()
	defer service.users.RUnlock()
	if nonce != service.users.data["test"].Nonce {
		t.Fatalf("Expected nonce: %v, got: %v", service.users.data["test"].Nonce, nonce)
	}
}

func TestCreateAccount(t *testing.T) {
	service := &bank{
		users: &userMap{
			data: make(map[string]proto.User),
		},
		txs: &txMap{
			data: make(map[uint64]proto.Transaction),
		},
	}
	user := proto.User{}
	_, err := service.CreateAccount(ctx, user)
	if !errors.Is(err, ErrEmptyAccount) {
		t.Fatalf("Expected error: %v, got: %v", ErrEmptyAccount, err)
	}

	user.Account = "test"
	_, err = service.CreateAccount(ctx, user)
	if !errors.Is(err, ErrEmptyPwd) {
		t.Fatalf("Expected error: %v, got: %v", ErrEmptyPwd, err)
	}

	user.Password = "XXXXXXXX"
	user.Balance = -1
	_, err = service.CreateAccount(ctx, user)
	if !errors.Is(err, ErrNegativeBalance) {
		t.Fatalf("Expected error: %v, got: %v", ErrNegativeBalance, err)
	}

	user.Balance = 100
	info, err := service.CreateAccount(ctx, user)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if info.Nonce == "" {
		t.Fatalf("Expected nonce: not empty, got: %v", info.Nonce)
	}

	_, err = service.CreateAccount(ctx, user)
	if !errors.Is(err, ErrAccountExist) {
		t.Fatalf("Expected error: %v, got: %v", ErrAccountExist, err)
	}
}

func TestDeposit(t *testing.T) {
	service := &bank{
		users: &userMap{
			data: make(map[string]proto.User),
		},
		txs: &txMap{
			data: make(map[uint64]proto.Transaction),
		},
		count:  1,
		search: NewSearch(),
	}

	tx := proto.Transaction{
		From: "test",
	}

	_, _, err := service.Deposit(ctx, tx, "")
	if !errors.Is(err, ErrFromAccount) {
		t.Fatalf("Expected error: %v, got: %v", ErrFromAccount, err)
	}

	tx.From = ""
	_, _, err = service.Deposit(ctx, tx, "")
	if !errors.Is(err, ErrToAccount) {
		t.Fatalf("Expected error: %v, got: %v", ErrToAccount, err)
	}

	tx.To = "test"
	tx.From = ""
	_, _, err = service.Deposit(ctx, tx, "")
	if !errors.Is(err, ErrNegativeBalance) {
		t.Fatalf("Expected error: %v, got: %v", ErrNegativeBalance, err)
	}

	tx.Amount = 10
	_, _, err = service.Deposit(ctx, tx, "")
	if !errors.Is(err, ErrEmptyNonce) {
		t.Fatalf("Expected error: %v, got: %v", ErrEmptyNonce, err)
	}

	service.users.Lock()
	service.users.data["test"] = proto.User{
		Account:  "test",
		Balance:  100,
		Password: "test-pwd",
		Name:     "test-user",
		Nonce:    "test-nonce",
	}
	service.users.Unlock()

	_, _, err = service.Deposit(ctx, tx, "test-no-nonce")
	if !errors.Is(err, ErrVerify) {
		t.Fatalf("Expected error: %v, got: %v", ErrVerify, err)
	}

	txLog, newNonce, err := service.Deposit(ctx, tx, "test-nonce")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if newNonce == "" || newNonce == "test-nonce" {
		t.Fatalf("Expected nonce: not empty, got: %v", newNonce)
	}
	if txLog.ID == 0 {
		t.Fatalf("Expected tx log: not empty, got: %v", txLog.ID)
	}
}

func TestWithdraw(t *testing.T) {
	service := &bank{
		users: &userMap{
			data: make(map[string]proto.User),
		},
		txs: &txMap{
			data: make(map[uint64]proto.Transaction),
		},
		count:  1,
		search: NewSearch(),
	}

	tx := proto.Transaction{
		From: "",
		To:   "test",
	}

	_, _, err := service.Withdraw(ctx, tx, "")
	if !errors.Is(err, ErrFromAccount) {
		t.Fatalf("Expected error: %v, got: %v", ErrFromAccount, err)
	}

	tx.From = "test"
	_, _, err = service.Withdraw(ctx, tx, "")
	if !errors.Is(err, ErrToAccount) {
		t.Fatalf("Expected error: %v, got: %v", ErrToAccount, err)
	}

	tx.To = ""
	_, _, err = service.Withdraw(ctx, tx, "")
	if !errors.Is(err, ErrNegativeBalance) {
		t.Fatalf("Expected error: %v, got: %v", ErrNegativeBalance, err)
	}

	tx.Amount = 1000
	_, _, err = service.Withdraw(ctx, tx, "")
	if !errors.Is(err, ErrEmptyNonce) {
		t.Fatalf("Expected error: %v, got: %v", ErrEmptyNonce, err)
	}

	service.users.Lock()
	service.users.data["test"] = proto.User{
		Account:  "test",
		Balance:  100,
		Password: "test-pwd",
		Name:     "test-user",
		Nonce:    "test-nonce",
	}
	service.users.Unlock()

	_, _, err = service.Withdraw(ctx, tx, "test-no-nonce")
	if !errors.Is(err, ErrVerify) {
		t.Fatalf("Expected error: %v, got: %v", ErrVerify, err)
	}

	_, _, err = service.Withdraw(ctx, tx, "test-nonce")
	if !errors.Is(err, ErrBalanceNotEnough) {
		t.Fatalf("Expected error: %v, got: %v", ErrBalanceNotEnough, err)
	}

	tx.Amount = 10

	txLog, newNonce, err := service.Withdraw(ctx, tx, "test-nonce")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if newNonce == "" || newNonce == "test-nonce" {
		t.Fatalf("Expected nonce: not empty, got: %v", newNonce)
	}
	if txLog.ID == 0 {
		t.Fatalf("Expected tx log: not empty, got: %v", txLog.ID)
	}
}

func TestTransaction(t *testing.T) {
	service := &bank{
		users: &userMap{
			data: make(map[string]proto.User),
		},
		txs: &txMap{
			data: make(map[uint64]proto.Transaction),
		},
		count:  1,
		search: NewSearch(),
	}

	tx := proto.Transaction{
		From: "",
		To:   "",
	}

	_, _, err := service.Transaction(ctx, tx, "")
	if !errors.Is(err, ErrFromAccount) {
		t.Fatalf("Expected error: %v, got: %v", ErrFromAccount, err)
	}

	tx.From = "test"
	_, _, err = service.Transaction(ctx, tx, "")
	if !errors.Is(err, ErrToAccount) {
		t.Fatalf("Expected error: %v, got: %v", ErrToAccount, err)
	}

	tx.To = "test2"
	_, _, err = service.Transaction(ctx, tx, "")
	if !errors.Is(err, ErrNegativeBalance) {
		t.Fatalf("Expected error: %v, got: %v", ErrNegativeBalance, err)
	}

	tx.Amount = 1000
	_, _, err = service.Transaction(ctx, tx, "")
	if !errors.Is(err, ErrEmptyNonce) {
		t.Fatalf("Expected error: %v, got: %v", ErrEmptyNonce, err)
	}

	service.users.Lock()
	service.users.data["test"] = proto.User{
		Account:  "test",
		Balance:  100,
		Password: "test-pwd",
		Name:     "test-user",
		Nonce:    "test-nonce",
	}

	service.users.data["test2"] = proto.User{
		Account:  "test2",
		Balance:  10,
		Password: "test2-pwd",
		Name:     "test2-user",
		Nonce:    "test2-nonce",
	}
	service.users.Unlock()

	_, _, err = service.Transaction(ctx, tx, "test-no-nonce")
	if !errors.Is(err, ErrVerify) {
		t.Fatalf("Expected error: %v, got: %v", ErrVerify, err)
	}

	_, _, err = service.Transaction(ctx, tx, "test-nonce")
	if !errors.Is(err, ErrBalanceNotEnough) {
		t.Fatalf("Expected error: %v, got: %v", ErrBalanceNotEnough, err)
	}

	tx.Amount = 10

	txLog, newNonce, err := service.Transaction(ctx, tx, "test-nonce")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if newNonce == "" || newNonce == "test-nonce" {
		t.Fatalf("Expected nonce: not empty, got: %v", newNonce)
	}
	if txLog.ID == 0 {
		t.Fatalf("Expected tx log: not empty, got: %v", txLog.ID)
	}
}

func TestGetBalance(t *testing.T) {
	service := &bank{
		users: &userMap{
			data: make(map[string]proto.User),
		},
		txs: &txMap{
			data: make(map[uint64]proto.Transaction),
		},
	}

	_, err := service.GetBalance(ctx, "")
	if !errors.Is(err, ErrEmptyAccount) {
		t.Fatalf("Expected error: %v, got: %v", ErrEmptyAccount, err)
	}
	_, err = service.GetBalance(ctx, "t")
	if !errors.Is(err, ErrAccountNotExist) {
		t.Fatalf("Expected error: %v, got: %v", ErrAccountNotExist, err)
	}

	service.users.Lock()
	service.users.data["test"] = proto.User{
		Account:  "test",
		Balance:  100,
		Password: "XXXXXXXX",
		Name:     "test-user",
		Nonce:    "test-nonce",
	}
	service.users.Unlock()

	balance, err := service.GetBalance(ctx, "test")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if balance != 100 {
		t.Fatalf("Expected balance: 100, got: %v", balance)
	}
}

func TestGetTransactions(t *testing.T) {
	service := &bank{
		users: &userMap{
			data: make(map[string]proto.User),
		},
		txs: &txMap{
			data: make(map[uint64]proto.Transaction),
		},
		count:  1,
		search: NewSearch(),
	}

	_, err := service.GetTransactions(ctx, "")
	if !errors.Is(err, ErrEmptyAccount) {
		t.Fatalf("Expected error: %v, got: %v", ErrEmptyAccount, err)
	}
	_, err = service.GetTransactions(ctx, "t")
	if !errors.Is(err, ErrAccountNotExist) {
		t.Fatalf("Expected error: %v, got: %v", ErrAccountNotExist, err)
	}
	pwd := uuid.NewString()

	service.users.Lock()
	service.users.data["test"] = proto.User{
		Account:  "test",
		Balance:  100,
		Password: pwd,
		Name:     "test-user",
		Nonce:    "test-nonce",
	}
	service.users.data["test2"] = proto.User{
		Account:  "test2",
		Balance:  102,
		Password: pwd,
		Name:     "test2-user",
		Nonce:    "test2-nonce",
	}
	service.users.Unlock()
	nonce, err := service.GetNonce(ctx, "test", pwd)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	_, nonce, err = service.Withdraw(ctx, proto.Transaction{
		From:   "test",
		Amount: 30,
	}, nonce)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	_, nonce, err = service.Deposit(ctx, proto.Transaction{
		To:     "test",
		Amount: 35,
	}, nonce)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	_, _, err = service.Transaction(ctx, proto.Transaction{
		From:   "test",
		To:     "test2",
		Amount: 10,
	}, nonce)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	txs, err := service.GetTransactions(ctx, "test")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(txs) != 3 {
		t.Fatalf("Expected txs: 3, got: %v", len(txs))
	}
}
