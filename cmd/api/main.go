package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"flipay/configs"
	"flipay/internal/auth"
	"flipay/internal/database"
	"flipay/internal/middleware"
	"flipay/internal/payment"
	"flipay/internal/queue"
	"flipay/internal/utils"
	"flipay/internal/webhook"
	"flipay/internal/worker"

	_ "flipay/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// @title Flipay - Mini Payment Gateway API
// @version 1.0
// @description Fintech backend simulation with JWT auth, payment processing, Redis worker, and webhook callbacks.
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg := configs.Load()
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.NewPostgres(rootCtx, cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("connect postgres", zap.Error(err))
	}
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
	worker.NewPaymentWorker(paymentRepo, paymentQueue, webhookSender, logger).Start(rootCtx)

	r := gin.New()
	r.Use(gin.Recovery(), middleware.RequestID(), middleware.RequestLogger(logger))

	// @Summary Health check
	// @Description Returns API health status for uptime monitoring and deployment health checks.
	// @Tags Health
	// @Produce json
	// @Success 200 {object} map[string]string
	// @Router /health [get]
	r.GET("/", func(c *gin.Context) { c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(utils.LandingPageHTML)) })
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v1")
	authRoutes := v1.Group("/auth", middleware.RateLimiter(redisClient, "rl:auth", 10, time.Minute))
	authRoutes.POST("/register", authHandler.Register)
	authRoutes.POST("/login", authHandler.Login)
	v1.POST("/callbacks/payment", webhookHandler.PaymentCallback)

	// Example of protected route group: every route below requires Authorization: Bearer <token>.
	protected := v1.Group("", middleware.JWT(cfg.JWTSecret), middleware.RateLimiter(redisClient, "rl:payment", 60, time.Minute))
	protected.POST("/payments", paymentHandler.Create)
	protected.GET("/payments/history", paymentHandler.History)
	protected.GET("/payments/:id", paymentHandler.Detail)

	server := &http.Server{Addr: ":" + cfg.AppPort, Handler: r, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		logger.Info("server started", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	<-rootCtx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", zap.Error(err))
	}
	logger.Info("server stopped gracefully")
}
