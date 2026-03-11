package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/openclaw/clawtool/internal/core"
	"github.com/spf13/cobra"
)

func newLogsCommand(localize func(string, map[string]any) string) *cobra.Command {
	var tail int
	var since string
	var bundle bool

	command := &cobra.Command{
		Use:   "logs",
		Short: localize("cmd.logs.short", nil),
		RunE: func(cmd *cobra.Command, _ []string) error {
			runtime := runtimeFromContext(cmd.Context())

			var sinceValue *time.Time
			if strings.TrimSpace(since) != "" {
				parsed, err := time.Parse(time.RFC3339, since)
				if err != nil {
					return err
				}
				sinceValue = &parsed
			}

			result, err := runtime.Service.Logs(tail, sinceValue, bundle)
			if err != nil {
				return err
			}

			return renderResult(cmd.Context(), cmd.OutOrStdout(), core.CommandResult{
				Command:    "logs",
				Status:     core.ResultStatusOK,
				SummaryKey: "logs.summary.ready",
				Details: []core.DetailItem{
					{Key: "log_path", Value: result.LogPath},
					{Key: "log_entries", Value: intString(len(result.Entries))},
					{Key: "tail", Value: intString(tail)},
					{Key: "since", Value: blankAsDash(since)},
					{Key: "bundle_path", Value: bundlePath(result.Bundle)},
					{Key: "entries", Value: formatLogEntries(result.Entries)},
				},
			})
		},
	}

	command.Flags().IntVar(&tail, "tail", 20, localize("flag.logs.tail", nil))
	command.Flags().StringVar(&since, "since", "", localize("flag.logs.since", nil))
	command.Flags().BoolVar(&bundle, "bundle", false, localize("flag.logs.bundle", nil))
	return command
}

func formatLogEntries(entries []core.LifecycleLogEntry) string {
	if len(entries) == 0 {
		return "-"
	}

	items := make([]string, 0, len(entries))
	for _, entry := range entries {
		items = append(items, fmt.Sprintf("%s:%s:%s", entry.Command, entry.Step, entry.Result))
	}
	return strings.Join(items, ", ")
}

func bundlePath(metadata *core.LogBundleMetadata) string {
	if metadata == nil {
		return "-"
	}
	return metadata.BundlePath
}
