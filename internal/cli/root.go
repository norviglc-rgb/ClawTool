package cli

import (
	"context"
	"os"

	"github.com/openclaw/clawtool/internal/app"
	"github.com/openclaw/clawtool/internal/config"
	clawi18n "github.com/openclaw/clawtool/internal/i18n"
	"github.com/openclaw/clawtool/internal/render"

	"github.com/spf13/cobra"
)

// Options stores root command flags. / Options 保存根命令标志。
type Options struct {
	Lang string
	JSON bool
}

// Execute runs the root command. / Execute 执行根命令。
func Execute() error {
	return NewRootCommand().Execute()
}

// NewRootCommand builds the Cobra tree. / NewRootCommand 构建 Cobra 命令树。
func NewRootCommand() *cobra.Command {
	options := &Options{}
	loader := clawi18n.Loader{}
	bundle, err := loader.NewBundle()
	if err != nil {
		panic(err)
	}

	defaultLocalizer := loader.NewLocalizer(bundle, "en")
	defaultText := localizeFunc(defaultLocalizer)

	cmd := &cobra.Command{
		Use:   "clawtool",
		Short: defaultText("root.short", nil),
		Long:  defaultText("root.long", nil),
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			resolver := clawi18n.Resolver{}
			cfg := config.RuntimeConfig{}
			lang := resolver.Resolve(options.Lang, cfg)
			localizer := loader.NewLocalizer(bundle, lang)
			localize := localizeFunc(localizer)

			var rendererImpl render.Renderer = render.HumanRenderer{Localize: localize}
			if options.JSON {
				rendererImpl = render.JSONRenderer{}
			}

			workingDir, err := os.Getwd()
			if err != nil {
				return err
			}

			runtime := Runtime{
				Config:   config.RuntimeConfig{Language: lang},
				Localize: localize,
				Renderer: rendererImpl,
				Service:  app.NewService(workingDir),
			}

			cmd.SetContext(withRuntime(context.Background(), runtime))
			cmd.SilenceUsage = true
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&options.Lang, "lang", "", defaultText("flag.lang", nil))
	cmd.PersistentFlags().BoolVar(&options.JSON, "json", false, defaultText("flag.json", nil))

	cmd.AddCommand(newDetectCommand(defaultText))
	cmd.AddCommand(newDoctorCommand(defaultText))
	cmd.AddCommand(newInitCommand(defaultText))
	cmd.AddCommand(newProfileCommand(defaultText))
	cmd.AddCommand(newPlanCommand(defaultText))
	cmd.AddCommand(newShowCommand(defaultText))
	cmd.AddCommand(newApplyCommand(defaultText))
	cmd.AddCommand(newVerifyCommand(defaultText))
	cmd.AddCommand(newInspectCommand(defaultText))
	cmd.AddCommand(newStatusCommand(defaultText))
	cmd.AddCommand(newLogsCommand(defaultText))
	cmd.AddCommand(newRollbackCommand(defaultText))
	cmd.AddCommand(newRepairCommand(defaultText))

	return cmd
}
