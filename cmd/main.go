package main

import (
	"TestHitalent2/internal/app"
	"TestHitalent2/internal/config"
	"TestHitalent2/pkg/logger"
	"context"
)

func main() {
	ctx := context.Background()
	cfg, err := config.NewConfig()
	if err != nil {
		panic(err)
	}
	ctx, err = logger.New(ctx)
	if err != nil {
		panic(err)
	}
	newApp := app.NewApp(cfg, ctx)
	newApp.MustRun()
}
