package cli

import (
	"encoding/json"
	"os"

	"github.com/openclaw/clawtool/internal/core"
	"github.com/spf13/cobra"
)

func newShowCommand(localize func(string, map[string]any) string) *cobra.Command {
	return &cobra.Command{
		Use:   "show <plan-file>",
		Short: localize("cmd.show.short", nil),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}

			var plan core.Plan
			if err := json.Unmarshal(data, &plan); err != nil {
				return err
			}

			return renderResult(cmd.Context(), cmd.OutOrStdout(), core.CommandResult{
				Command:    "show",
				Status:     core.ResultStatusOK,
				SummaryKey: "show.summary.ready",
				Details: []core.DetailItem{
					{Key: "plan_file", Value: args[0]},
					{Key: "profile", Value: plan.Profile},
					{Key: "generated_config", Value: plan.GeneratedConfig},
					{Key: "requires_backup", Value: boolString(plan.RequiresBackup)},
					{Key: "verification_steps", Value: joinOrDash(plan.VerificationSteps)},
					{Key: "plan_steps", Value: intString(len(plan.Steps))},
					{Key: "changes", Value: formatPlanChanges(plan.Changes)},
					{Key: "content_diffs", Value: formatContentDiffs(plan.ContentDiffs)},
				},
			})
		},
	}
}
