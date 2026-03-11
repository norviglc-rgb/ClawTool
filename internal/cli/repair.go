package cli

import (
	"fmt"
	"strings"

	"github.com/openclaw/clawtool/internal/core"
	"github.com/spf13/cobra"
)

func newRepairCommand(localize func(string, map[string]any) string) *cobra.Command {
	var applySafe bool
	var yes bool

	command := &cobra.Command{
		Use:   "repair",
		Short: localize("cmd.repair.short", nil),
		RunE: func(cmd *cobra.Command, _ []string) error {
			runtime := runtimeFromContext(cmd.Context())
			result, err := runtime.Service.Repair(applySafe, yes)
			if err != nil {
				return err
			}

			return renderResult(cmd.Context(), cmd.OutOrStdout(), core.CommandResult{
				Command:    "repair",
				Status:     core.ResultStatusOK,
				SummaryKey: "repair.summary.ready",
				Details: []core.DetailItem{
					{Key: "repair_actions", Value: formatRepairActions(runtime.Localize, result.Actions)},
					{Key: "applied_count", Value: intString(result.AppliedCount)},
					{Key: "apply_safe", Value: boolString(applySafe)},
					{Key: "yes", Value: boolString(yes)},
				},
			})
		},
	}

	command.Flags().BoolVar(&applySafe, "apply-safe", false, localize("flag.repair.apply_safe", nil))
	command.Flags().BoolVar(&yes, "yes", false, localize("flag.yes", nil))
	return command
}

func formatRepairActions(localize func(string, map[string]any) string, actions []core.RepairAction) string {
	if len(actions) == 0 {
		return "-"
	}

	items := make([]string, 0, len(actions))
	for _, action := range actions {
		flags := "-"
		if strings.TrimSpace(action.RequiresFlag) != "" {
			flags = action.RequiresFlag
		}
		items = append(items, fmt.Sprintf("%s|risky=%t|applied=%t|flag=%s", localize(action.MessageKey, nil), action.Risky, action.Applied, flags))
	}
	return strings.Join(items, ", ")
}
