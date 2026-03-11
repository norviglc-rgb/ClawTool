package cli

import (
	"github.com/openclaw/clawtool/internal/core"
	"github.com/spf13/cobra"
)

func newStatusCommand(localize func(string, map[string]any) string) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: localize("cmd.status.short", nil),
		RunE: func(cmd *cobra.Command, _ []string) error {
			runtime := runtimeFromContext(cmd.Context())
			result, err := runtime.Service.Status()
			if err != nil {
				return err
			}

			lastApply := "-"
			if result.LastApplyAt != nil {
				lastApply = result.LastApplyAt.UTC().Format("2006-01-02T15:04:05Z")
			}
			lastRollback := "-"
			if result.LastRollbackAt != nil {
				lastRollback = result.LastRollbackAt.UTC().Format("2006-01-02T15:04:05Z")
			}

			return renderResult(cmd.Context(), cmd.OutOrStdout(), core.CommandResult{
				Command:    "status",
				Status:     core.ResultStatusOK,
				SummaryKey: "status.summary.ready",
				Details: []core.DetailItem{
					{Key: "profile", Value: blankAsDash(result.CurrentProfile)},
					{Key: "last_apply_at", Value: lastApply},
					{Key: "last_apply_result", Value: blankAsDash(result.LastApplyResult)},
					{Key: "last_rollback_at", Value: lastRollback},
					{Key: "last_rollback_result", Value: blankAsDash(result.LastRollbackResult)},
					{Key: "backup_count", Value: intString(result.BackupCount)},
				},
			})
		},
	}
}
