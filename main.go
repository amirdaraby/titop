package main

import (
	"context"

	"github.com/amirdaraby/titop/internal/titop"
)

func main() {

	ctx := context.Background()

	if err := titop.Run(ctx); err != nil {
		panic(err)
	}
}
