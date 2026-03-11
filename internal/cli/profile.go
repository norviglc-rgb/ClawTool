package cli

import (
	"fmt"
	"strings"

	"github.com/openclaw/clawtool/internal/core"
	"github.com/spf13/cobra"
)

func newProfileCommand(localize func(string, map[string]any) string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: localize("cmd.profile.short", nil),
	}

	cmd.AddCommand(newProfileListCommand(localize))
	cmd.AddCommand(newProfileShowCommand(localize))
	cmd.AddCommand(newProfileCreateCommand(localize))
	cmd.AddCommand(newProfileValidateCommand(localize))
	cmd.AddCommand(newProfileUseCommand(localize))

	return cmd
}

func newProfileListCommand(localize func(string, map[string]any) string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: localize("cmd.profile.list.short", nil),
		RunE: func(cmd *cobra.Command, _ []string) error {
			runtime := runtimeFromContext(cmd.Context())
			result, err := runtime.Service.ListProfiles()
			if err != nil {
				return err
			}

			items := make([]string, 0, len(result.Profiles))
			for _, profile := range result.Profiles {
				status := "inactive"
				if profile.Active {
					status = "active"
				}
				items = append(items, fmt.Sprintf("%s (%s)", profile.Name, status))
			}
			details := []core.DetailItem{}
			if len(items) > 0 {
				details = append(details, core.DetailItem{Key: "profiles", Value: strings.Join(items, ", ")})
			}

			return renderResult(cmd.Context(), cmd.OutOrStdout(), core.CommandResult{
				Command:    "profile list",
				Status:     core.ResultStatusOK,
				SummaryKey: "profile.list.summary",
				Details:    details,
			})
		},
	}
}

func newProfileShowCommand(localize func(string, map[string]any) string) *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: localize("cmd.profile.show.short", nil),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime := runtimeFromContext(cmd.Context())
			result, err := runtime.Service.ShowProfile(args[0])
			if err != nil {
				return err
			}

			return renderResult(cmd.Context(), cmd.OutOrStdout(), core.CommandResult{
				Command:    "profile show",
				Status:     core.ResultStatusOK,
				SummaryKey: "profile.show.summary",
				Details: []core.DetailItem{
					{Key: "name", Value: result.Profile.Name},
					{Key: "target_kind", Value: result.Profile.Target.Kind},
					{Key: "target_address", Value: blankAsDash(result.Profile.Target.Address)},
					{Key: "target_user", Value: blankAsDash(result.Profile.Target.User)},
					{Key: "target_port", Value: blankAsDash(zeroIntAsBlank(result.Profile.Target.Port))},
					{Key: "target_key_path", Value: blankAsDash(result.Profile.Target.KeyPath)},
					{Key: "host_key_strategy", Value: blankAsDash(result.Profile.Target.HostKeyStrategy)},
					{Key: "path", Value: result.Path},
					{Key: "active", Value: boolString(result.Active)},
				},
			})
		},
	}
}

func newProfileCreateCommand(localize func(string, map[string]any) string) *cobra.Command {
	var kind string
	var address string
	var user string
	var port int
	var keyPath string
	var hostKeyStrategy string

	command := &cobra.Command{
		Use:   "create <name>",
		Short: localize("cmd.profile.create.short", nil),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime := runtimeFromContext(cmd.Context())
			result, err := runtime.Service.CreateProfile(core.Profile{
				Version: "v1",
				Name:    args[0],
				Target: core.ProfileTarget{
					Kind:            kind,
					Address:         address,
					User:            user,
					Port:            port,
					KeyPath:         keyPath,
					HostKeyStrategy: hostKeyStrategy,
				},
			})
			if err != nil {
				return err
			}

			return renderResult(cmd.Context(), cmd.OutOrStdout(), core.CommandResult{
				Command:    "profile create",
				Status:     core.ResultStatusOK,
				SummaryKey: "profile.create.summary",
				Details: []core.DetailItem{
					{Key: "name", Value: result.Profile.Name},
					{Key: "target_kind", Value: result.Profile.Target.Kind},
					{Key: "target_address", Value: blankAsDash(result.Profile.Target.Address)},
					{Key: "target_user", Value: blankAsDash(result.Profile.Target.User)},
					{Key: "target_port", Value: blankAsDash(zeroIntAsBlank(result.Profile.Target.Port))},
					{Key: "target_key_path", Value: blankAsDash(result.Profile.Target.KeyPath)},
					{Key: "host_key_strategy", Value: blankAsDash(result.Profile.Target.HostKeyStrategy)},
					{Key: "path", Value: result.Path},
				},
			})
		},
	}

	command.Flags().StringVar(&kind, "kind", "local", localize("flag.profile.kind", nil))
	command.Flags().StringVar(&address, "address", "", localize("flag.profile.address", nil))
	command.Flags().StringVar(&user, "user", "", localize("flag.profile.user", nil))
	command.Flags().IntVar(&port, "port", 0, localize("flag.profile.port", nil))
	command.Flags().StringVar(&keyPath, "key-path", "", localize("flag.profile.key_path", nil))
	command.Flags().StringVar(&hostKeyStrategy, "host-key-strategy", "", localize("flag.profile.host_key_strategy", nil))
	return command
}

func newProfileValidateCommand(localize func(string, map[string]any) string) *cobra.Command {
	return &cobra.Command{
		Use:   "validate <name>",
		Short: localize("cmd.profile.validate.short", nil),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime := runtimeFromContext(cmd.Context())
			result, err := runtime.Service.ValidateProfile(args[0])
			if err != nil {
				return err
			}

			return renderResult(cmd.Context(), cmd.OutOrStdout(), core.CommandResult{
				Command:    "profile validate",
				Status:     core.ResultStatusOK,
				SummaryKey: "profile.validate.summary",
				Details: []core.DetailItem{
					{Key: "name", Value: result.Profile.Name},
					{Key: "path", Value: result.Path},
				},
			})
		},
	}
}

func newProfileUseCommand(localize func(string, map[string]any) string) *cobra.Command {
	return &cobra.Command{
		Use:   "use <name>",
		Short: localize("cmd.profile.use.short", nil),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime := runtimeFromContext(cmd.Context())
			result, err := runtime.Service.UseProfile(args[0])
			if err != nil {
				return err
			}

			return renderResult(cmd.Context(), cmd.OutOrStdout(), core.CommandResult{
				Command:    "profile use",
				Status:     core.ResultStatusOK,
				SummaryKey: "profile.use.summary",
				Details: []core.DetailItem{
					{Key: "name", Value: result.Name},
					{Key: "path", Value: result.Path},
				},
			})
		},
	}
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func intString(value int) string {
	return fmt.Sprintf("%d", value)
}

func zeroIntAsBlank(value int) string {
	if value == 0 {
		return ""
	}
	return intString(value)
}

func blankAsDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func filepathBaseOrValue(value string) string {
	if value == "" {
		return "-"
	}
	parts := strings.Split(strings.ReplaceAll(value, "\\", "/"), "/")
	return parts[len(parts)-1]
}

func joinOrDash(values []string) string {
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ", ")
}
