package cli

import (
	glazedcli "github.com/go-go-golems/glazed/pkg/cli"
	"github.com/spf13/cobra"
)

func NewCommands() ([]*cobra.Command, error) {
	serveCmd, err := NewServeCommand()
	if err != nil {
		return nil, err
	}

	cobraServeCmd, err := glazedcli.BuildCobraCommandFromCommand(
		serveCmd,
		glazedcli.WithParserConfig(glazedcli.CobraParserConfig{
			AppName: "markdown_quizz",
		}),
	)
	if err != nil {
		return nil, err
	}

	return []*cobra.Command{cobraServeCmd}, nil
}
