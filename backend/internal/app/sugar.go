package app

import (
	"fmt"
	"os"

	"github.com/benbjohnson/clock"
	"github.com/leshless/golibrary/interrupt"
)

func Run() {
	primitives := NewPrimitives(
		clock.New(),
		interrupt.NewInterrupter(),
		os.DirFS("/"),
	)

	app, err := InitApp(primitives)
	if err != nil {
		panic(fmt.Errorf("Failed to initialize app: %w", err))
	}

	if err = app.Run(); err != nil {
		panic(fmt.Errorf("App finished with error: %w", err))
	}
}
