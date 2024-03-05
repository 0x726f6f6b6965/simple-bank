package router

import (
	"github.com/0x726f6f6b6965/bank/internal/api"
	"github.com/0x726f6f6b6965/bank/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(server *gin.Engine) {
	RegisterUserRouter(server.Group("/account"))
	RegisterBankRouter(server.Group("/bank"))
}

func RegisterBankRouter(router *gin.RouterGroup) {
	router.Use(middleware.UserAuthorization())
	router.GET("/balance", api.BankAPI.GetBalance)
	router.POST("/transfer", api.BankAPI.Transfer)
	router.GET("/transactions", api.BankAPI.GetTransactions)
}

func RegisterUserRouter(router *gin.RouterGroup) {
	router.POST("/nonce", api.BankAPI.GetToken)
	router.POST("/register", api.BankAPI.CreateAccount)
}
