package main

import (
	"context"
	"Practice3/internal/app"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	appInstance := app.NewApp(ctx)
	appInstance.Run()
}