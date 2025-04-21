package main

import (
	"context"

	titop "github.com/amirdaraby/titop/internal/application"
	"github.com/amirdaraby/titop/internal/shared"
)

func main() {

	if err := shared.Init(); err != nil {
		panic(err)
	}

	ctx := context.Background()

	if err := titop.Run(ctx); err != nil {
		panic(err)
	}
}
