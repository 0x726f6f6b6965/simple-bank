package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/0x726f6f6b6965/bank/internal/api/router"
	"github.com/0x726f6f6b6965/bank/internal/api/services"
	"github.com/0x726f6f6b6965/bank/internal/proto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	ctx         context.Context
	baseURL     string
	contentType = "application/json"
	client      *http.Client
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	ctx = context.Background()
	services.NewBank()
	engin := gin.Default()
	gin.SetMode(gin.TestMode)
	router.RegisterRoutes(engin)
	go engin.Run(":8080")
	baseURL = "http://localhost:8080"
	client = &http.Client{}
	fmt.Printf("\033[1;33m%s\033[0m", "> Setup completed\n")
}

func teardown() {
	fmt.Printf("\033[1;33m%s\033[0m", "> Teardown completed")
	fmt.Printf("\n")
}

func TestCreateAccount(t *testing.T) {
	pwd := uuid.NewString()
	req := &proto.CreateAccountRequest{
		Password: pwd,
		Name:     uuid.NewString(),
		Balance:  100,
	}
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Post(fmt.Sprintf("%s/account/register", baseURL), contentType, bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
	result := make(map[string]interface{})
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if result["code"].(float64) != 200 {
		t.Fatalf("expected code %s, got %s", "200", result["code"])
	}
	if result["message"].(string) != "success" {
		t.Fatalf("expected message %s, got %s", "success", result["message"])
	}
	if result["data"].(map[string]interface{})["account"].(string) == "" {
		t.Fatalf("expected account not empty")
	}

	if result["data"].(map[string]interface{})["name"].(string) != req.Name {
		t.Fatalf("expected name %s, got %s", req.Name, result["data"].(map[string]interface{})["name"])
	}
}

func TestGetToken(t *testing.T) {
	// register
	pwd := uuid.NewString()
	user, err := register(pwd, 203)
	if err != nil {
		t.Fatal(err)
	}

	//get token
	reqGetToken := &proto.GetTokenRequest{
		Account:  user.Account,
		Password: pwd,
	}
	b, err := json.Marshal(reqGetToken)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Post(fmt.Sprintf("%s/account/nonce", baseURL), contentType, bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
	result := make(map[string]interface{})
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if result["code"].(float64) != 200 {
		t.Fatalf("expected code %s, got %s", "200", result["code"])
	}
	if result["message"].(string) != "success" {
		t.Fatalf("expected message %s, got %s", "success", result["message"])
	}
	if result["data"].(string) == "" {
		t.Fatalf("expected data not empty")
	}
}

func TestGetBalance(t *testing.T) {
	// register
	pwd := uuid.NewString()
	user, err := register(pwd, 203)
	if err != nil {
		t.Fatal(err)
	}

	// get token
	token, err := getToken(user.Account, pwd)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/bank/balance", baseURL), nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
	result := make(map[string]interface{})
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if result["code"].(float64) != 200 {
		t.Fatalf("expected code %s, got %s", "200", result["code"])
	}
	if result["message"].(string) != "success" {
		t.Fatalf("expected message %s, got %s", "success", result["message"])
	}
	if int(result["data"].(float64)) != user.Balance {
		t.Fatalf("expected data %d, got %s", user.Balance, result["data"])
	}
}

func TestDeposit(t *testing.T) {
	// register
	pwd := uuid.NewString()
	user, err := register(pwd, 203)
	if err != nil {
		t.Fatal(err)
	}

	// get token
	token, err := getToken(user.Account, pwd)
	if err != nil {
		t.Fatal(err)
	}

	// deposit
	body := &proto.TransactionRequest{
		Action: proto.TransactionActionDeposit,
		To:     user.Account,
		Amount: 100,
	}
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/bank/transfer", baseURL), bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
	result := make(map[string]interface{})
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result["code"].(float64) != 200 {
		t.Fatalf("expected code %s, got %s", "200", result["code"])
	}
}

func TestWithdraw(t *testing.T) {
	// register
	pwd := uuid.NewString()
	user, err := register(pwd, 203)
	if err != nil {
		t.Fatal(err)
	}

	// get token
	token, err := getToken(user.Account, pwd)
	if err != nil {
		t.Fatal(err)
	}

	// withdraw
	body := &proto.TransactionRequest{
		Action: proto.TransactionActionWithdraw,
		From:   user.Account,
		Amount: 100,
	}
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/bank/transfer", baseURL), bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
	result := make(map[string]interface{})
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result["code"].(float64) != 200 {
		t.Fatalf("expected code %s, got %s", "200", result["code"])
	}
}

func TestTransaction(t *testing.T) {
	// register
	pwd := uuid.NewString()
	from, err := register(pwd, 203)
	if err != nil {
		t.Fatal(err)
	}

	pwd2 := uuid.NewString()
	to, err := register(pwd2, 10)
	if err != nil {
		t.Fatal(err)
	}

	// get token
	token, err := getToken(from.Account, pwd)
	if err != nil {
		t.Fatal(err)
	}

	// transaction
	body := &proto.TransactionRequest{
		Action: proto.TransactionActionTransfer,
		From:   from.Account,
		To:     to.Account,
		Amount: 100,
	}
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/bank/transfer", baseURL), bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
	result := make(map[string]interface{})
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result["code"].(float64) != 200 {
		t.Fatalf("expected code %s, got %s", "200", result["code"])
	}
}

func TestGetTransactions(t *testing.T) {
	// register
	pwd := uuid.NewString()
	user, err := register(pwd, 203)
	if err != nil {
		t.Fatal(err)
	}

	// register2
	pwd2 := uuid.NewString()
	user2, err := register(pwd2, 305)
	if err != nil {
		t.Fatal(err)
	}

	// get token
	token, err := getToken(user.Account, pwd)
	if err != nil {
		t.Fatal(err)
	}

	// get token2
	token2, err := getToken(user2.Account, pwd2)
	if err != nil {
		t.Fatal(err)
	}

	// transaction1
	body := &proto.TransactionRequest{
		Action: proto.TransactionActionTransfer,
		From:   user.Account,
		To:     user2.Account,
		Amount: 100,
	}

	b, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/bank/transfer", baseURL), bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	_, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	// transaction2
	body = &proto.TransactionRequest{
		Action: proto.TransactionActionDeposit,
		To:     user.Account,
		Amount: 100,
	}

	token, err = getToken(user.Account, pwd)
	if err != nil {
		t.Fatal(err)
	}

	b, err = json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("%s/bank/transfer", baseURL), bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	_, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	// transaction3
	body = &proto.TransactionRequest{
		Action: proto.TransactionActionWithdraw,
		From:   user.Account,
		Amount: 10,
	}

	token, err = getToken(user.Account, pwd)
	if err != nil {
		t.Fatal(err)
	}

	b, err = json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("%s/bank/transfer", baseURL), bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	_, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	// get transactions
	token, err = getToken(user.Account, pwd)
	if err != nil {
		t.Fatal(err)
	}
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/bank/transactions", baseURL), nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	result := make(map[string]interface{})
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result["code"].(float64) != 200 {
		t.Fatalf("expected code %s, got %s", "200", result["code"])
	}
	txs1 := result["data"].([]interface{})
	if len(txs1) != 3 {
		t.Fatalf("expected 3 transactions, got %d", len(txs1))
	}

	// get transactions2
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/bank/transactions", baseURL), nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token2))
	resp, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	result = make(map[string]interface{})
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result["code"].(float64) != 200 {
		t.Fatalf("expected code %s, got %s", "200", result["code"])
	}
	txs2 := result["data"].([]interface{})
	if len(txs2) != 1 {
		t.Fatalf("expected 3 transactions, got %d", len(txs2))
	}
}

func register(pwd string, balance int) (*proto.User, error) {

	req := &proto.CreateAccountRequest{
		Password: pwd,
		Name:     uuid.NewString(),
		Balance:  balance,
	}
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(fmt.Sprintf("%s/account/register", baseURL), contentType, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	user := new(proto.User)
	// convert map to json
	b, err = json.Marshal(result["data"].(map[string]interface{}))
	if err != nil {
		return nil, err
	}

	// convert json to struct
	err = json.Unmarshal(b, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func getToken(account, pwd string) (string, error) {
	reqGetNonce := &proto.GetTokenRequest{
		Account:  account,
		Password: pwd,
	}
	b, err := json.Marshal(reqGetNonce)
	if err != nil {
		return "", err
	}
	resp, err := http.Post(fmt.Sprintf("%s/account/nonce", baseURL), contentType, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	result := make(map[string]interface{})
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result["data"].(string), nil
}
