package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openclaw/clawtool/internal/core"
	"github.com/spf13/cobra"
)

func newPlanCommand(localize func(string, map[string]any) string) *cobra.Command {
	var outPath string

	command := &cobra.Command{
		Use:   "plan",
		Short: localize("cmd.plan.short", nil),
		RunE: func(cmd *cobra.Command, _ []string) error {
			runtime := runtimeFromContext(cmd.Context())
			plan, err := runtime.Service.Plan()
			if err != nil {
				return err
			}
			if strings.TrimSpace(outPath) != "" {
				if err := savePlan(plan, outPath); err != nil {
					return err
				}
			}

			return renderResult(cmd.Context(), cmd.OutOrStdout(), core.CommandResult{
				Command:    "plan",
				Status:     core.ResultStatusOK,
				SummaryKey: "plan.summary.ready",
				Details: []core.DetailItem{
					{Key: "plan_file", Value: blankAsDash(outPath)},
					{Key: "profile", Value: plan.Profile},
					{Key: "generated_config", Value: plan.GeneratedConfig},
					{Key: "requires_backup", Value: boolString(plan.RequiresBackup)},
					{Key: "verification_steps", Value: strings.Join(plan.VerificationSteps, ", ")},
					{Key: "plan_steps", Value: intString(len(plan.Steps))},
					{Key: "changes", Value: formatPlanChanges(plan.Changes)},
					{Key: "content_diffs", Value: formatContentDiffs(plan.ContentDiffs)},
				},
			})
		},
	}

	command.Flags().StringVar(&outPath, "out", "", localize("flag.plan.out", nil))
	return command
}

func formatPlanChanges(changes []core.PlanChange) string {
	items := make([]string, 0, len(changes))
	for _, change := range changes {
		items = append(items, fmt.Sprintf("%s=%s", change.Action, change.Path))
	}
	return strings.Join(items, ", ")
}

func formatContentDiffs(diffs []core.ContentDiff) string {
	if len(diffs) == 0 {
		return "-"
	}

	items := make([]string, 0, len(diffs))
	for _, diff := range diffs {
		switch diff.Action {
		case "add":
			items = append(items, fmt.Sprintf("add:%s=%s", diff.Field, diff.After))
		case "remove":
			items = append(items, fmt.Sprintf("remove:%s=%s", diff.Field, diff.Before))
		default:
			items = append(items, fmt.Sprintf("update:%s=%s->%s", diff.Field, diff.Before, diff.After))
		}
	}
	return strings.Join(items, ", ")
}

func savePlan(plan core.Plan, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}
