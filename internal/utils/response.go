package utils

import (
	"time"

	"github.com/gin-gonic/gin"
)

// Response - Unified return response interface format
func Response(ctx *gin.Context, code int, errCode int, errMsg string, data interface{}) {
	ctx.JSON(code, map[string]interface{}{
		"code":        errCode,
		"currentTime": time.Now().UnixMilli(),
		"message":     errMsg,
		"data":        data,
	})
}
