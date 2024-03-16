package application

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

// public
type App struct {
	router http.Handler
	rdb    *redis.Client
}

func New() *App {
	app := &App{
		rdb: redis.NewClient(&redis.Options{}),
	}

	app.loadRoutes()

	return app
}

// method for App, to start the server, not return it tho
func (app *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":8081",
		Handler: app.router,
	}

	if err := app.rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	defer func() {
		if err := app.rdb.Close(); err != nil {
			fmt.Println("Faild to Close redis:", err)
		}
	}()

	fmt.Println("Starting server on port 8081")
	channel := make(chan error, 1)

	// to run server concurrently
	// iffe?
	go func() {
		var err = server.ListenAndServe()
		if err != nil {
			// [Can't do this since it's a goroutine] ->  return fmt.Errorf("failed to start server: %w", err)
			// NOT channel <- err
			channel <- fmt.Errorf("failed to start server: %w", err)
		}

		close(channel)
	}()

	// to listen for 2 channel simulatenously if one of this is blocked
	select {
	case err := <-channel:
		return err

	case <-ctx.Done():
		// brand new context for server.shutdown(here context)
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		return server.Shutdown(timeout)
	}
	// return nil
}
