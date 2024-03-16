package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/raymondreddd/golnag/application"
)

func main() {
	app := application.New()

	/*
		Graceful shutdown
		// context.Background() as parent context, and we listen for only signit (close)
		// we are getting new derived context one ctx
		another is a cnacellation function, which is linked to derived context(a nd its children)
	*/

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// todo is passing all parameters: context.TODO()
	err := app.Start(ctx)
	if err != nil {
		fmt.Println("failed to start app:", err)
	}
}
