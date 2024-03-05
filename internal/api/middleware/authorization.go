package middleware

import (
	"net/http"

	"github.com/0x726f6f6b6965/bank/internal/utils"
	"github.com/gin-gonic/gin"
)

func UserAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := utils.ExtractTokenMetadata(c.Request)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set("access_token", token)
		c.Next()
	}
}
