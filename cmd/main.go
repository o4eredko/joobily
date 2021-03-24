package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"gitlab.jooble.com/marketing_tech/joobily/config"
	"gitlab.jooble.com/marketing_tech/joobily/infrastructure/api"
	"gitlab.jooble.com/marketing_tech/joobily/infrastructure/app"
)

func main() {
	c := config.NewConfig("config.toml")
	app := app.New(c)

	api := api.New(c, app)
	// Start server
	go api.Start()
	// Wait for interrupt signal to gracefully shutdown the server with
	// api timeout of 10 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	api.Shutdown(ctx)
}
