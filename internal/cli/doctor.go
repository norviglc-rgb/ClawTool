package cli

import (
	"fmt"

	"github.com/openclaw/clawtool/internal/core"
	"github.com/spf13/cobra"
)

func newDoctorCommand(localize func(string, map[string]any) string) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: localize("cmd.doctor.short", nil),
		RunE: func(cmd *cobra.Command, _ []string) error {
			runtime := runtimeFromContext(cmd.Context())
			result, err := runtime.Service.Doctor()
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

			details := make([]core.DetailItem, 0, len(result.Findings))
			for _, finding := range result.Findings {
				message := runtime.Localize(finding.MessageKey, nil)
				details = append(details, core.DetailItem{
					Key:   finding.Code,
					Value: fmt.Sprintf("%s %s", finding.Severity, message),
				})
			}

			return renderResult(cmd.Context(), cmd.OutOrStdout(), core.CommandResult{
				Command:    "doctor",
				Status:     status,
				SummaryKey: "doctor.summary.ready",
				Details:    details,
			})
		},
	}
}
