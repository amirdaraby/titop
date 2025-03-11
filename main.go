package main

import (
	"context"

	titop "github.com/amirdaraby/titop/internal/application"
)

func main() {

	ctx := context.Background()

	if err := titop.Run(ctx); err != nil {
		panic(err)
	}
}
