package app

import (
	"TestHitalent2/internal/config"
	"TestHitalent2/internal/repository"
	"TestHitalent2/internal/service"
	"TestHitalent2/internal/transport"
	"TestHitalent2/pkg/logger"
	"TestHitalent2/pkg/postgres"
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type App struct {
	OrganizationServer *transport.OrganizationServer
	cfg                *config.Config
	ctx                context.Context
	wg                 sync.WaitGroup
	cancel             context.CancelFunc
}

func NewApp(cfg *config.Config, context context.Context) *App {
	db, err := postgres.New(cfg.Postgres)
	if err != nil {
		panic(err)
	}

	if err := runMigrations(db, context); err != nil {
		panic(err)
	}

	repo := repository.NewOrganizationRepository(db, context)
	srv := service.NewOrganizationService(context, repo)
	server := transport.NewOrganizationServer(cfg, srv, context)
	return &App{
		OrganizationServer: server,
		cfg:                cfg,
		ctx:                context,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	errCh := make(chan error, 1)
	a.wg.Add(1)
	go func() {
		logger.GetLoggerFromCtx(a.ctx).Info("Server started on address", zap.Any("address", a.cfg.Host+":"+a.cfg.Port))
		defer a.wg.Done()
		if err := a.OrganizationServer.Run(); err != nil {
			errCh <- err
			a.cancel()
		}
	}()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-errCh:
		logger.GetLoggerFromCtx(a.ctx).Error("error running app", zap.Error(err))
		return err
	case <-a.ctx.Done():
		logger.GetLoggerFromCtx(a.ctx).Info("context done")
	}

	return nil
}

func runMigrations(db *gorm.DB, ctx context.Context) error {
	logger.GetLoggerFromCtx(ctx).Info("Running database migrations...")

	sqlDB, err := db.DB()
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error("Failed to get database instance", zap.Error(err))
		return err
	}

	if err := goose.SetDialect("postgres"); err != nil {
		logger.GetLoggerFromCtx(ctx).Error("Failed to set goose dialect", zap.Error(err))
		return err
	}

	if err := goose.Up(sqlDB, "migrations"); err != nil {
		logger.GetLoggerFromCtx(ctx).Error("Migration failed", zap.Error(err))
		return err
	}

	logger.GetLoggerFromCtx(ctx).Info("Database migrations completed successfully!")
	return nil
}
