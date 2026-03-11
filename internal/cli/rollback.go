package cli

import (
	"github.com/openclaw/clawtool/internal/core"
	"github.com/spf13/cobra"
)

func newRollbackCommand(localize func(string, map[string]any) string) *cobra.Command {
	return &cobra.Command{
		Use:   "rollback [backup-id]",
		Short: localize("cmd.rollback.short", nil),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime := runtimeFromContext(cmd.Context())
			backupID := ""
			if len(args) == 1 {
				backupID = args[0]
			}

			result, err := runtime.Service.Rollback(backupID)
			if err != nil {
				return err
			}

			return renderResult(cmd.Context(), cmd.OutOrStdout(), core.CommandResult{
				Command:    "rollback",
				Status:     core.ResultStatusOK,
				SummaryKey: "rollback.summary.ready",
				Details: []core.DetailItem{
					{Key: "selected_backup", Value: result.SelectedBackup.ID},
					{Key: "restored_profile", Value: result.RestoredProfile},
					{Key: "restored_path", Value: result.RestoredPath},
					{Key: "pre_rollback_backup_path", Value: blankAsDash(result.PreRollbackBackupPath)},
					{Key: "state_path", Value: result.StatePath},
					{Key: "verify_findings", Value: intString(len(result.VerifyResult.Findings))},
				},
			})
		},
	}
}
