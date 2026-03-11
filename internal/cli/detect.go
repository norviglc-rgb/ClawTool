package cli

import (
	"github.com/openclaw/clawtool/internal/core"
	"github.com/spf13/cobra"
)

func newDetectCommand(localize func(string, map[string]any) string) *cobra.Command {
	return &cobra.Command{
		Use:   "detect",
		Short: localize("cmd.detect.short", nil),
		RunE: func(cmd *cobra.Command, _ []string) error {
			runtime := runtimeFromContext(cmd.Context())
			result, err := runtime.Service.Detect()
			if err != nil {
				return err
			}

			return renderResult(cmd.Context(), cmd.OutOrStdout(), core.CommandResult{
				Command:    "detect",
				Status:     core.ResultStatusOK,
				SummaryKey: "detect.summary.ready",
				Details: []core.DetailItem{
					{Key: "os", Value: result.OS},
					{Key: "arch", Value: result.Arch},
					{Key: "working_dir", Value: result.WorkingDir},
					{Key: "openclaw_installed", Value: boolString(result.OpenClawInstalled)},
				},
			})
		},
	}
}
