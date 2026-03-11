package cli

import (
	"strings"

	"github.com/openclaw/clawtool/internal/core"
	"github.com/spf13/cobra"
)

func newInspectCommand(localize func(string, map[string]any) string) *cobra.Command {
	return &cobra.Command{
		Use:   "inspect",
		Short: localize("cmd.inspect.short", nil),
		RunE: func(cmd *cobra.Command, _ []string) error {
			runtime := runtimeFromContext(cmd.Context())
			result, err := runtime.Service.Inspect()
			if err != nil {
				return err
			}

			profiles := make([]string, 0, len(result.Profiles))
			for _, profile := range result.Profiles {
				state := "inactive"
				if profile.Active {
					state = "active"
				}
				profiles = append(profiles, profile.Name+" ("+state+")")
			}
			lastApply := "-"
			if result.LastApplyAt != nil {
				lastApply = result.LastApplyAt.UTC().Format("2006-01-02T15:04:05Z")
			}
			backups := make([]string, 0, len(result.Backups))
			for _, backup := range result.Backups {
				backups = append(backups, backup.ID)
			}

			return renderResult(cmd.Context(), cmd.OutOrStdout(), core.CommandResult{
				Command:    "inspect",
				Status:     core.ResultStatusOK,
				SummaryKey: "inspect.summary.ready",
				Details: []core.DetailItem{
					{Key: "profile", Value: blankAsDash(result.CurrentProfile)},
					{Key: "profiles", Value: strings.Join(profiles, ", ")},
					{Key: "install_path", Value: blankAsDash(result.InstallPath)},
					{Key: "config_path", Value: blankAsDash(result.ConfigPath)},
					{Key: "state_path", Value: result.StatePath},
					{Key: "plan_record_path", Value: result.PlanRecordPath},
					{Key: "last_apply_at", Value: lastApply},
					{Key: "last_apply_result", Value: blankAsDash(result.LastApplyResult)},
					{Key: "backup_count", Value: intString(len(result.Backups))},
					{Key: "backups", Value: blankAsDash(strings.Join(backups, ", "))},
				},
			})
		},
	}
}
