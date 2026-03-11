package cli

import (
	"strings"

	"github.com/openclaw/clawtool/internal/core"
	"github.com/spf13/cobra"
)

func newInitCommand(localize func(string, map[string]any) string) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: localize("cmd.init.short", nil),
		RunE: func(cmd *cobra.Command, _ []string) error {
			runtime := runtimeFromContext(cmd.Context())
			result, err := runtime.Service.Init()
			if err != nil {
				return err
			}

			return renderResult(cmd.Context(), cmd.OutOrStdout(), core.CommandResult{
				Command:    "init",
				Status:     core.ResultStatusOK,
				SummaryKey: "init.summary.ready",
				Details: []core.DetailItem{
					{Key: "workspace", Value: result.WorkspacePath},
					{Key: "created_count", Value: intString(len(result.CreatedPaths))},
					{Key: "existing_count", Value: intString(len(result.ExistingPaths))},
					{Key: "default_profile", Value: filepathBaseOrValue(result.DefaultProfile)},
					{Key: "created_paths", Value: strings.Join(result.CreatedPaths, ", ")},
				},
			})
		},
	}
}
