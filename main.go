package main

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/exp/slog"
)

func main() {
	ctx := context.Background()
	token, err := retrieveToken(ctx)
	if err != nil {
		slog.Error("error retrieving token", err)
		os.Exit(1)
	}

	fmt.Println(token.AccessToken)
}
