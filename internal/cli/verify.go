package cli

import (
	"fmt"

	"github.com/openclaw/clawtool/internal/core"
	"github.com/spf13/cobra"
)

func newVerifyCommand(localize func(string, map[string]any) string) *cobra.Command {
	return &cobra.Command{
		Use:   "verify",
		Short: localize("cmd.verify.short", nil),
		RunE: func(cmd *cobra.Command, _ []string) error {
			runtime := runtimeFromContext(cmd.Context())
			result, err := runtime.Service.Verify()
			if err != nil {
				return err
			}

			status := core.ResultStatusOK
			for _, finding := range result.Findings {
				switch finding.Severity {
				case core.SeverityFail:
					status = core.ResultStatusError
				case core.SeverityWarn:
					if status != core.ResultStatusError {
						status = core.ResultStatusWarn
					}
				}
			}

			details := []core.DetailItem{{Key: "profile", Value: result.Profile}}
			for _, finding := range result.Findings {
				details = append(details, core.DetailItem{
					Key:   "finding",
					Value: fmt.Sprintf("%s %s", finding.Severity, runtime.Localize(finding.MessageKey, nil)),
				})
			}

			return renderResult(cmd.Context(), cmd.OutOrStdout(), core.CommandResult{
				Command:    "verify",
				Status:     status,
				SummaryKey: "verify.summary.ready",
				Details:    details,
			})
		},
	}
}
