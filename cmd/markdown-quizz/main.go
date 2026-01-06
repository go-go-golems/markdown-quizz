package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/XXX/internal/cli"
	"github.com/spf13/cobra"
)

func buildRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "markdown-quizz",
		Short: "markdown-quizz backend (Go port)",
	}
	return root
}

func main() {
	root := buildRoot()

	commands, err := cli.NewCommands()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error creating commands: %v\n", err)
		os.Exit(1)
	}

	for _, c := range commands {
		root.AddCommand(c)
	}

	if err := root.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
