package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/openclaw/clawtool/internal/cli"
	"github.com/openclaw/clawtool/internal/core"
)

func main() {
	if err := cli.Execute(); err != nil {
		var exitErr *core.ExitError
		if errors.As(err, &exitErr) {
			if !exitErr.Silent && exitErr.Cause != nil {
				fmt.Fprintln(os.Stderr, exitErr.Cause)
			}
			os.Exit(exitErr.Code)
		}

		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
