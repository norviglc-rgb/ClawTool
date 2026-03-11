package cli

import (
	"strings"

	"github.com/openclaw/clawtool/internal/core"
	"github.com/spf13/cobra"
)

func newRemoteCommand(localize func(string, map[string]any) string) *cobra.Command {
	command := &cobra.Command{
		Use:   "remote",
		Short: localize("cmd.remote.short", nil),
	}

	command.AddCommand(newRemotePlanCommand(localize))
	command.AddCommand(newRemoteApplyCommand(localize))
	command.AddCommand(newRemoteVerifyCommand(localize))
	command.AddCommand(newRemoteExecCommand(localize))
	return command
}

func newRemotePlanCommand(localize func(string, map[string]any) string) *cobra.Command {
	return &cobra.Command{
		Use:   "plan <profile>",
		Short: localize("cmd.remote.plan.short", nil),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime := runtimeFromContext(cmd.Context())
			profile, _, err := runtime.Remote.LoadProfile(args[0])
			if err != nil {
				return err
			}

			plan := runtime.Remote.Plan(profile)
			return renderResult(cmd.Context(), cmd.OutOrStdout(), core.CommandResult{
				Command:    "remote plan",
				Status:     core.ResultStatusOK,
				SummaryKey: "remote.plan.summary.ready",
				Details: []core.DetailItem{
					{Key: "profile", Value: profile.Name},
					{Key: "target_address", Value: profile.Target.Address},
					{Key: "requires_backup", Value: boolString(plan.RequiresBackup)},
					{Key: "verification_steps", Value: strings.Join(plan.VerificationSteps, ", ")},
					{Key: "plan_steps", Value: intString(len(plan.Steps))},
					{Key: "plan_step_details", Value: formatPlanSteps(runtime.Localize, plan.Steps)},
					{Key: "changes", Value: formatPlanChanges(plan.Changes)},
				},
			})
		},
	}
}

func newRemoteVerifyCommand(localize func(string, map[string]any) string) *cobra.Command {
	return &cobra.Command{
		Use:   "verify <profile>",
		Short: localize("cmd.remote.verify.short", nil),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime := runtimeFromContext(cmd.Context())
			profile, _, err := runtime.Remote.LoadProfile(args[0])
			if err != nil {
				return err
			}

			result := runtime.Remote.Verify(profile)
			status := applyCommandStatus(result.Findings)
			details := []core.DetailItem{
				{Key: "profile", Value: profile.Name},
				{Key: "target_address", Value: profile.Target.Address},
			}
			for _, finding := range result.Findings {
				details = append(details, core.DetailItem{
					Key:   "finding",
					Value: findingString(runtime.Localize, finding),
				})
			}

			commandResult := core.CommandResult{
				Command:    "remote verify",
				Status:     status,
				SummaryKey: "remote.verify.summary.ready",
				Details:    details,
			}
			if err := renderResult(cmd.Context(), cmd.OutOrStdout(), commandResult); err != nil {
				return err
			}
			if status == core.ResultStatusError {
				return &core.ExitError{Code: 1, Silent: true}
			}
			return nil
		},
	}
}

func newRemoteApplyCommand(localize func(string, map[string]any) string) *cobra.Command {
	return &cobra.Command{
		Use:   "apply <profile>",
		Short: localize("cmd.remote.apply.short", nil),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime := runtimeFromContext(cmd.Context())
			profile, _, err := runtime.Remote.LoadProfile(args[0])
			if err != nil {
				return err
			}

			result, err := runtime.Remote.Apply(cmd.Context(), profile)
			if err != nil {
				return err
			}

			status := applyCommandStatus(result.VerifyResult.Findings)
			commandResult := core.CommandResult{
				Command:    "remote apply",
				Status:     status,
				SummaryKey: "remote.apply.summary.ready",
				Details: []core.DetailItem{
					{Key: "profile", Value: result.Plan.Profile},
					{Key: "target_address", Value: profile.Target.Address},
					{Key: "remote_config_path", Value: result.RemoteConfigPath},
					{Key: "generated_config", Value: result.GeneratedConfig},
					{Key: "plan_record_path", Value: result.PlanRecordPath},
					{Key: "state_path", Value: result.StatePath},
					{Key: "changed", Value: boolString(result.Changed)},
					{Key: "backup_path", Value: blankAsDash(result.BackupPath)},
					{Key: "verify_result", Value: verifyResultValue(result.VerifyResult.Findings)},
				},
			}
			if err := renderResult(cmd.Context(), cmd.OutOrStdout(), commandResult); err != nil {
				return err
			}
			if status == core.ResultStatusError {
				return &core.ExitError{Code: 1, Silent: true}
			}
			return nil
		},
	}
}

func newRemoteExecCommand(localize func(string, map[string]any) string) *cobra.Command {
	return &cobra.Command{
		Use:   "exec <profile> <command...>",
		Short: localize("cmd.remote.exec.short", nil),
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime := runtimeFromContext(cmd.Context())
			profile, _, err := runtime.Remote.LoadProfile(args[0])
			if err != nil {
				return err
			}

			commandLine := strings.Join(args[1:], " ")
			result, err := runtime.Remote.Exec(cmd.Context(), profile, commandLine)
			if err != nil {
				return err
			}

			status := core.ResultStatusOK
			if result.ExitCode != 0 {
				status = core.ResultStatusError
			}

			commandResult := core.CommandResult{
				Command:    "remote exec",
				Status:     status,
				SummaryKey: "remote.exec.summary.ready",
				Details: []core.DetailItem{
					{Key: "profile", Value: result.Profile},
					{Key: "target_address", Value: result.Address},
					{Key: "target_user", Value: result.User},
					{Key: "target_port", Value: intString(result.Port)},
					{Key: "host_key_strategy", Value: result.HostKeyStrategy},
					{Key: "command", Value: result.Command},
					{Key: "exit_code", Value: intString(result.ExitCode)},
					{Key: "stdout", Value: blankAsDash(result.Stdout)},
					{Key: "stderr", Value: blankAsDash(result.Stderr)},
				},
			}
			if err := renderResult(cmd.Context(), cmd.OutOrStdout(), commandResult); err != nil {
				return err
			}
			if result.ExitCode != 0 {
				return &core.ExitError{Code: result.ExitCode, Silent: true}
			}
			return nil
		},
	}
}
