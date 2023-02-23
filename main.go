package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"golang.org/x/exp/slog"
)

var (
	version = pflag.BoolP("version", "V", false, "Print version number")
)

func main() {
	pflag.Parse()

	if *version {
		if err := printVersion(); err != nil {
			slog.Error("error retrieving version information", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	ctx := context.Background()
	token, err := retrieveToken(ctx)
	if err != nil {
		slog.Error("error retrieving token", err)
		os.Exit(1)
	}

	fmt.Println(token.AccessToken)
}
