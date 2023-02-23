package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"golang.org/x/exp/slog"
)

var (
	deb     = pflag.BoolP("debug", "d", false, "Turn on debugging")
	version = pflag.BoolP("version", "V", false, "Print version number")
)

func main() {
	pflag.Parse()

	ctx := context.Background()

	if *deb {
		setDebugLevel(ctx)
	}

	if *version {
		if err := printVersion(); err != nil {
			slog.Error("error retrieving version information", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	token, err := retrieveToken(ctx)
	if err != nil {
		slog.Error("error retrieving token", err)
		os.Exit(1)
	}

	slog.Log(slog.LevelDebug, "minted", slog.String("token", fmt.Sprintf("%#v", token)))

	fmt.Println(token.AccessToken)
}

func setDebugLevel(ctx context.Context) {
	programLevel := new(slog.LevelVar)
	programLevel.Set(slog.LevelDebug)
	h := slog.HandlerOptions{Level: programLevel}.NewTextHandler(os.Stderr)
	slog.SetDefault(slog.New(h).WithContext(ctx))
}
