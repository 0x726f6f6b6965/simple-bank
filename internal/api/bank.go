package api

import (
	"net/http"
	"time"

	"github.com/0x726f6f6b6965/bank/internal/api/services"
	"github.com/0x726f6f6b6965/bank/internal/proto"
	"github.com/0x726f6f6b6965/bank/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var BankAPI *bankApi

type bankApi struct{}

func (api *bankApi) GetBalance(ctx *gin.Context) {
	b := services.GetBankService()
	if b == nil {
		utils.Response(ctx, http.StatusOK, http.StatusInternalServerError, "service not found", nil)
		return
	}
	var param *proto.UserToken
	if token, ok := ctx.Get("access_token"); !ok || token.(*proto.UserToken) == nil {
		utils.Response(ctx, http.StatusOK, http.StatusBadRequest, "invalid token", nil)
		return
	} else {
		param = token.(*proto.UserToken)
	}

	balance, err := b.GetBalance(ctx, param.Account)
	if err != nil {
		utils.Response(ctx, http.StatusOK, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.Response(ctx, http.StatusOK, http.StatusOK, "success", balance)
}

func (api *bankApi) Transfer(ctx *gin.Context) {
	b := services.GetBankService()
	if b == nil {
		utils.Response(ctx, http.StatusOK, http.StatusInternalServerError, "service not found", nil)
		return
	}

	var token *proto.UserToken
	if t, ok := ctx.Get("access_token"); !ok || t.(*proto.UserToken) == nil {
		utils.Response(ctx, http.StatusOK, http.StatusBadRequest, "invalid token", nil)
		return
	} else {
		token = t.(*proto.UserToken)
	}

	var param proto.TransactionRequest
	if err := ctx.ShouldBindJSON(&param); err != nil {
		utils.Response(ctx, http.StatusOK, http.StatusBadRequest, err.Error(), nil)
		return
	}
	tx := proto.Transaction{
		From:   param.From,
		To:     param.To,
		Amount: param.Amount,
	}

	var (
		result *proto.Transaction
		err    error
	)

	switch param.Action {
	case proto.TransactionActionDeposit:
		result, _, err = b.Deposit(ctx, tx, token.Nonce)
	case proto.TransactionActionWithdraw:
		result, _, err = b.Withdraw(ctx, tx, token.Nonce)
	case proto.TransactionActionTransfer:
		result, _, err = b.Transaction(ctx, tx, token.Nonce)
	default:
		utils.Response(ctx, http.StatusOK, http.StatusBadRequest, "invalid action", nil)
		return
	}
	if err != nil {
		utils.Response(ctx, http.StatusOK, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp := proto.TransactionResponse{
		ID: result.ID,
	}
	utils.Response(ctx, http.StatusOK, http.StatusOK, "success", resp)
}

func (api *bankApi) GetToken(ctx *gin.Context) {
	b := services.GetBankService()
	if b == nil {
		utils.Response(ctx, http.StatusOK, http.StatusInternalServerError, "service not found", nil)
		return
	}
	var param proto.GetTokenRequest
	if err := ctx.ShouldBindJSON(&param); err != nil {
		utils.Response(ctx, http.StatusOK, http.StatusBadRequest, err.Error(), nil)
		return
	}
	nonce, err := b.GetNonce(ctx, param.Account, param.Password)
	if err != nil {
		utils.Response(ctx, http.StatusOK, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	token, _ := utils.GenerateNewAccessToken(param.Account, nonce, 5*time.Minute)
	utils.Response(ctx, http.StatusOK, http.StatusOK, "success", token)
}

func (api *bankApi) CreateAccount(ctx *gin.Context) {
	b := services.GetBankService()
	if b == nil {
		utils.Response(ctx, http.StatusOK, http.StatusInternalServerError, "service not found", nil)
		return
	}
	var param proto.CreateAccountRequest
	if err := ctx.ShouldBindJSON(&param); err != nil {
		utils.Response(ctx, http.StatusOK, http.StatusBadRequest, err.Error(), nil)
		return
	}
	user := proto.User{
		Account:  uuid.NewString(),
		Password: param.Password,
		Name:     param.Name,
		Balance:  param.Balance,
	}
	resp, err := b.CreateAccount(ctx, user)
	if err != nil {
		utils.Response(ctx, http.StatusOK, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp.Password = ""
	utils.Response(ctx, http.StatusOK, http.StatusOK, "success", resp)
}

func (api *bankApi) GetTransactions(ctx *gin.Context) {
	b := services.GetBankService()
	if b == nil {
		utils.Response(ctx, http.StatusOK, http.StatusInternalServerError, "service not found", nil)
		return
	}
	var param *proto.UserToken
	if token, ok := ctx.Get("access_token"); !ok || token.(*proto.UserToken) == nil {
		utils.Response(ctx, http.StatusOK, http.StatusBadRequest, "invalid token", nil)
		return
	} else {
		param = token.(*proto.UserToken)
	}
	resp, err := b.GetTransactions(ctx, param.Account)
	if err != nil {
		utils.Response(ctx, http.StatusOK, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.Response(ctx, http.StatusOK, http.StatusOK, "success", resp)
}
