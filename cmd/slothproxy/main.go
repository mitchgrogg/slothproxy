package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/mitchgrogg/rita-devtools/slothproxy/pkg/slothproxy"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	p := slothproxy.New()
	cmd := buildRootCommand(p)

	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}
