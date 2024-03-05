package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/0x726f6f6b6965/bank/internal/api/router"
	"github.com/0x726f6f6b6965/bank/internal/api/services"
	"github.com/0x726f6f6b6965/bank/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

func main() {
	godotenv.Load()
	path := os.Getenv("CONFIG")
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("read yaml error", err)
		return
	}

	var cfg config.AppConfig
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatal("unmarshal yaml error", err)
		return
	}

	services.NewBank()

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HttpPort),
		Handler: initEngine(&cfg),
	}

	if err := server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Println("Server was shutdown gracefully")
		} else {
			log.Fatal("Server error", err)
		}
	}
}

func initEngine(cfg *config.AppConfig) *gin.Engine {
	gin.SetMode(func() string {
		if cfg.IsDevEnv() {
			return gin.DebugMode
		}
		return gin.ReleaseMode
	}())
	engine := gin.New()
	engine.Use(gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  "Service internal exception!",
		})
	}))
	router.RegisterRoutes(engine)
	return engine
}
