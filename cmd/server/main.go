package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/TheTuxis/gondor-event-bus/internal/config"
	"github.com/TheTuxis/gondor-event-bus/internal/handler"
	"github.com/TheTuxis/gondor-event-bus/internal/middleware"
	"github.com/TheTuxis/gondor-event-bus/internal/model"
	natspkg "github.com/TheTuxis/gondor-event-bus/internal/nats"
	jwtpkg "github.com/TheTuxis/gondor-event-bus/internal/pkg/jwt"
	"github.com/TheTuxis/gondor-event-bus/internal/repository"
	"github.com/TheTuxis/gondor-event-bus/internal/service"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Init logger
	var logger *zap.Logger
	var err error
	if cfg.Environment == "production" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = logger.Sync() }()

	// Connect to PostgreSQL
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal("failed to get underlying sql.DB", zap.Error(err))
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Auto-migrate models
	if err := db.AutoMigrate(
		&model.Webhook{},
		&model.DeadLetterMessage{},
		&model.EventLog{},
	); err != nil {
		logger.Fatal("failed to auto-migrate", zap.Error(err))
	}
	logger.Info("database migration completed")

	// Init Redis client
	var redisClient *redis.Client
	if cfg.RedisURL != "" {
		opts, err := redis.ParseURL("redis://" + cfg.RedisURL)
		if err != nil {
			// Fallback: treat as host:port
			opts = &redis.Options{Addr: cfg.RedisURL}
		}
		redisClient = redis.NewClient(opts)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			logger.Warn("redis connection failed, continuing without redis", zap.Error(err))
			redisClient = nil
		} else {
			logger.Info("connected to Redis")
		}
	}

	// Connect to NATS JetStream
	var natsClient *natspkg.Client
	if cfg.NATSURL != "" {
		natsClient, err = natspkg.Connect(cfg.NATSURL, logger)
		if err != nil {
			logger.Warn("NATS connection failed, continuing without NATS", zap.Error(err))
			natsClient = nil
		}
	}

	// Init JWT manager (validate-only — tokens are issued by gondor-users-security)
	jwtManager := jwtpkg.NewManager(cfg.JWTSecret)

	// Init repositories
	webhookRepo := repository.NewWebhookRepository(db)
	eventLogRepo := repository.NewEventLogRepository(db)
	dlqRepo := repository.NewDeadLetterRepository(db)

	// Init services
	eventService := service.NewEventService(eventLogRepo, natsClient, logger)
	webhookService := service.NewWebhookService(webhookRepo, logger)
	dlqService := service.NewDeadLetterService(dlqRepo, logger)

	// Init handlers
	healthHandler := handler.NewHealthHandler(db, redisClient)
	eventHandler := handler.NewEventHandler(eventService)
	webhookHandler := handler.NewWebhookHandler(webhookService)
	dlqHandler := handler.NewDeadLetterHandler(dlqService)

	// Setup Gin
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.AuthMiddleware(jwtManager))

	// Health & metrics (no auth required — handled by skip list)
	router.GET("/health", healthHandler.Health)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Event routes
	v1 := router.Group("/v1")
	{
		// Publish & log
		v1.POST("/events/publish", eventHandler.Publish)
		v1.GET("/events/log", eventHandler.ListLogs)

		// Webhooks
		v1.GET("/events/webhooks", webhookHandler.List)
		v1.POST("/events/webhooks", webhookHandler.Create)
		v1.GET("/events/webhooks/:id", webhookHandler.GetByID)
		v1.PUT("/events/webhooks/:id", webhookHandler.Update)
		v1.DELETE("/events/webhooks/:id", webhookHandler.Delete)
		v1.POST("/events/webhooks/:id/test", webhookHandler.Test)

		// Dead letter queue
		v1.GET("/events/dlq", dlqHandler.List)
		v1.POST("/events/dlq/:id/retry", dlqHandler.Retry)
		v1.DELETE("/events/dlq/:id", dlqHandler.Delete)
	}

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("starting server", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}

	if natsClient != nil {
		natsClient.Close()
	}
	if redisClient != nil {
		_ = redisClient.Close()
	}
	_ = sqlDB.Close()

	logger.Info("server stopped")
}
