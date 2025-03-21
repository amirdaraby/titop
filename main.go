package main

import (
	"context"

	titop "github.com/amirdaraby/titop/internal/application"
	"github.com/amirdaraby/titop/internal/config"
)

func main() {

	if err := config.Init(); err != nil {
		panic(err)
	}

	ctx := context.Background()

	if err := titop.Run(ctx); err != nil {
		panic(err)
	}
}
