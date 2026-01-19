package cli

import (
	glazedcli "github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
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

	build := func(command cmds.Command) (*cobra.Command, error) {
		return glazedcli.BuildCobraCommandFromCommand(
			command,
			glazedcli.WithParserConfig(glazedcli.CobraParserConfig{
				AppName: "markdown_quizz",
			}),
		)
	}

	// documents
	documentsParent := &cobra.Command{
		Use:   "documents",
		Short: "Document operations",
	}
	docListCmd, err := NewDocumentsListCommand()
	if err != nil {
		return nil, err
	}
	cobraDocList, err := build(docListCmd)
	if err != nil {
		return nil, err
	}
	docImportCmd, err := NewDocumentsImportCommand()
	if err != nil {
		return nil, err
	}
	cobraDocImport, err := build(docImportCmd)
	if err != nil {
		return nil, err
	}
	docGetCmd, err := NewDocumentsGetCommand()
	if err != nil {
		return nil, err
	}
	cobraDocGet, err := build(docGetCmd)
	if err != nil {
		return nil, err
	}
	docUpdateCmd, err := NewDocumentsUpdateCommand()
	if err != nil {
		return nil, err
	}
	cobraDocUpdate, err := build(docUpdateCmd)
	if err != nil {
		return nil, err
	}
	docDeleteCmd, err := NewDocumentsDeleteCommand()
	if err != nil {
		return nil, err
	}
	cobraDocDelete, err := build(docDeleteCmd)
	if err != nil {
		return nil, err
	}
	docAnalyticsCmd, err := NewDocumentsAnalyticsCommand()
	if err != nil {
		return nil, err
	}
	cobraDocAnalytics, err := build(docAnalyticsCmd)
	if err != nil {
		return nil, err
	}
	docSubmissionsCmd, err := NewDocumentsSubmissionsCommand()
	if err != nil {
		return nil, err
	}
	cobraDocSubmissions, err := build(docSubmissionsCmd)
	if err != nil {
		return nil, err
	}

	documentsParent.AddCommand(cobraDocList, cobraDocImport, cobraDocGet, cobraDocUpdate, cobraDocDelete, cobraDocAnalytics, cobraDocSubmissions)

	// quiz
	quizParent := &cobra.Command{
		Use:   "quiz",
		Short: "Quiz operations",
	}
	quizSubmitCmd, err := NewQuizSubmitCommand()
	if err != nil {
		return nil, err
	}
	cobraQuizSubmit, err := build(quizSubmitCmd)
	if err != nil {
		return nil, err
	}
	quizSubmitBatchCmd, err := NewQuizSubmitBatchCommand()
	if err != nil {
		return nil, err
	}
	cobraQuizSubmitBatch, err := build(quizSubmitBatchCmd)
	if err != nil {
		return nil, err
	}
	quizParent.AddCommand(cobraQuizSubmit, cobraQuizSubmitBatch)

	// submissions
	submissionsParent := &cobra.Command{
		Use:   "submissions",
		Short: "Submission operations",
	}
	subMineCmd, err := NewSubmissionsMineCommand()
	if err != nil {
		return nil, err
	}
	cobraSubMine, err := build(subMineCmd)
	if err != nil {
		return nil, err
	}
	subByDocCmd, err := NewSubmissionsByDocumentCommand()
	if err != nil {
		return nil, err
	}
	cobraSubByDoc, err := build(subByDocCmd)
	if err != nil {
		return nil, err
	}
	subGetCmd, err := NewSubmissionsGetCommand()
	if err != nil {
		return nil, err
	}
	cobraSubGet, err := build(subGetCmd)
	if err != nil {
		return nil, err
	}
	submissionsParent.AddCommand(cobraSubMine, cobraSubByDoc, cobraSubGet)

	return []*cobra.Command{
		cobraServeCmd,
		documentsParent,
		quizParent,
		submissionsParent,
	}, nil
}
