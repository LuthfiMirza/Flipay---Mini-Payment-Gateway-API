package main

import (
	"context"
	"net/http"

	"flipay/configs"
	"flipay/internal/auth"
	"flipay/internal/database"
	"flipay/internal/middleware"
	"flipay/internal/payment"
	"flipay/internal/queue"
	"flipay/internal/webhook"
	"flipay/internal/worker"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	cfg := configs.Load()
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	db, err := database.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil { logger.Fatal("connect postgres", zap.Error(err)) }
	defer db.Close()

	redisClient := database.NewRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	defer redisClient.Close()

	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, cfg.JWTSecret, cfg.JWTTTLHours)
	authHandler := auth.NewHandler(authService)

	paymentRepo := payment.NewRepository(db)
	paymentQueue := queue.NewRedisQueue(redisClient)
	paymentService := payment.NewService(paymentRepo, paymentQueue, cfg.PaymentExpiryMinutes, logger)
	paymentHandler := payment.NewHandler(paymentService, logger)
	webhookSender := webhook.NewSender(cfg.WebhookURL, cfg.WebhookSecret, db, logger)
	webhookHandler := webhook.NewHandler(cfg.WebhookSecret, logger)
	worker.NewPaymentWorker(paymentRepo, paymentQueue, webhookSender, logger).Start(ctx)

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	v1 := r.Group("/api/v1")
	v1.POST("/auth/register", authHandler.Register)
	v1.POST("/auth/login", authHandler.Login)
	v1.POST("/callbacks/payment", webhookHandler.PaymentCallback)

	// Example of protected route group: every route below requires Authorization: Bearer <token>.
	protected := v1.Group("")
	protected.Use(middleware.JWT(cfg.JWTSecret))
	protected.POST("/payments", paymentHandler.Create)
	protected.GET("/payments/history", paymentHandler.History)
	protected.GET("/payments/:id", paymentHandler.Detail)

	logger.Fatal("server stopped", zap.Error(r.Run(":"+cfg.AppPort)))
}
