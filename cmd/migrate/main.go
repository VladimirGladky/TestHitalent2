package main

import (
	"TestHitalent2/internal/config"
	"TestHitalent2/pkg/logger"
	"TestHitalent2/pkg/postgres"
	"context"
	"os"

	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	ctx, err := logger.New(ctx)
	if err != nil {
		panic(err)
	}

	cfg, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	db, err := postgres.New(cfg.Postgres)
	if err != nil {
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error("Failed to get database instance", zap.Error(err))
		panic(err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		logger.GetLoggerFromCtx(ctx).Error("Failed to set goose dialect", zap.Error(err))
		panic(err)
	}

	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	logger.GetLoggerFromCtx(ctx).Info("Running migration command", zap.String("command", command))

	switch command {
	case "up":
		if err := goose.Up(sqlDB, "migrations"); err != nil {
			logger.GetLoggerFromCtx(ctx).Error("Migration up failed", zap.Error(err))
			panic(err)
		}
		logger.GetLoggerFromCtx(ctx).Info("Migrations applied successfully!")

	case "down":
		if err := goose.Down(sqlDB, "migrations"); err != nil {
			logger.GetLoggerFromCtx(ctx).Error("Migration down failed", zap.Error(err))
			panic(err)
		}
		logger.GetLoggerFromCtx(ctx).Info("Last migration rolled back!")

	case "reset":
		if err := goose.Reset(sqlDB, "migrations"); err != nil {
			logger.GetLoggerFromCtx(ctx).Error("Migration reset failed", zap.Error(err))
			panic(err)
		}
		logger.GetLoggerFromCtx(ctx).Info("All migrations rolled back!")

	case "status":
		if err := goose.Status(sqlDB, "migrations"); err != nil {
			logger.GetLoggerFromCtx(ctx).Error("Migration status failed", zap.Error(err))
			panic(err)
		}

	default:
		logger.GetLoggerFromCtx(ctx).Error("Unknown command", zap.String("command", command))
		os.Exit(1)
	}
}
