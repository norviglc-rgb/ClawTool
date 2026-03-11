package cli

import (
	"github.com/openclaw/clawtool/internal/core"
	"github.com/spf13/cobra"
)

func newApplyCommand(localize func(string, map[string]any) string) *cobra.Command {
	return &cobra.Command{
		Use:   "apply",
		Short: localize("cmd.apply.short", nil),
		RunE: func(cmd *cobra.Command, _ []string) error {
			runtime := runtimeFromContext(cmd.Context())
			result, err := runtime.Service.Apply()
			if err != nil {
				return err
			}

			return renderResult(cmd.Context(), cmd.OutOrStdout(), core.CommandResult{
				Command:    "apply",
				Status:     core.ResultStatusOK,
				SummaryKey: "apply.summary.ready",
				Details: []core.DetailItem{
					{Key: "profile", Value: result.Plan.Profile},
					{Key: "generated_config", Value: result.GeneratedConfig},
					{Key: "plan_record_path", Value: result.PlanRecordPath},
					{Key: "state_path", Value: result.StatePath},
					{Key: "changed", Value: boolString(result.Changed)},
					{Key: "backup_path", Value: blankAsDash(result.BackupPath)},
				},
			})
		},
	}
}
