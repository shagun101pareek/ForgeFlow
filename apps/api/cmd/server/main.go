package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/shagun101pareek/forgeflow/internal/cache"
	"github.com/shagun101pareek/forgeflow/internal/config"
	"github.com/shagun101pareek/forgeflow/internal/database"
	"github.com/shagun101pareek/forgeflow/internal/logger"
	"github.com/shagun101pareek/forgeflow/internal/routes"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.AppEnv)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	deps := routes.Dependencies{
		Log:    log,
		AppEnv: cfg.AppEnv,
	}

	db, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Warn().Err(err).Msg("postgres unavailable; starting in degraded mode")
	} else {
		deps.DB = db
		deps.DBReady = true
		defer db.Close()
	}

	redisClient, err := cache.NewRedisClient(ctx, cfg.RedisURL)
	if err != nil {
		log.Warn().Err(err).Msg("redis unavailable; starting in degraded mode")
	} else {
		deps.Redis = redisClient
		deps.RedisOK = true
		defer redisClient.Close()
	}

	app := fiber.New(fiber.Config{
		AppName:      "ForgeFlow API",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	})

	routes.Register(app, deps)

	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		log.Info().Str("addr", addr).Msg("starting forgeflow api")
		if err := app.Listen(addr); err != nil {
			log.Fatal().Err(err).Msg("fiber stopped")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("graceful shutdown failed")
	}
}
