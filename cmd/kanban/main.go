package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/shuntaka9576/kanban/cmd/kanban/cmd"
	"github.com/spf13/cobra"
)

func main() {
	hasDebug := os.Getenv("DEBUG") != ""

	if cmd, err := cmd.RootCmd.ExecuteC(); err != nil {
		printError(os.Stderr, err, cmd, hasDebug)
		os.Exit(1)
	}
}

func printError(out io.Writer, err error, command *cobra.Command, debug bool) {
	fmt.Fprintln(out, err)

	var flagError cmd.FlagError
	if errors.As(err, &flagError) || strings.HasPrefix(err.Error(), "unknown command ") {
		if !strings.HasSuffix(err.Error(), "\n") {
			fmt.Fprintln(out)
		}
		fmt.Fprintln(out, command.UsageString())
	}
}
