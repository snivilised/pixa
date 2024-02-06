package command

import (
	"fmt"
	"strings"

	"github.com/snivilised/cobrass/src/assistant"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type MagickParameterSet struct {
	// remove the follow:
	//
	Directory string
	Concise   bool
	Pattern   string
	Threshold uint
}

const MagickPsName = "magick-ps"

func (b *Bootstrap) buildMagickCommand(container *assistant.CobraContainer) *cobra.Command {
	magickCommand := &cobra.Command{
		Use:   "mag",
		Short: "mag (magick) sub command",
		Long:  "Long description of the magick command",
		RunE: func(cmd *cobra.Command, args []string) error {
			var appErr error

			ps := container.MustGetParamSet(MagickPsName).(*assistant.ParamSet[MagickParameterSet]) //nolint:errcheck // is Must call

			if err := ps.Validate(); err == nil {
				// --> native := ps.Native

				// optionally invoke cross field validation
				//
				if xv := ps.CrossValidate(func(ps *MagickParameterSet) error {
					return nil
				}); xv == nil {
					options := []string{}
					cmd.Flags().Visit(func(f *pflag.Flag) {
						options = append(options, fmt.Sprintf("--%v=%v", f.Name, f.Value))
					})

					fmt.Printf("%v %v Running magick, with options: '%v', args: '%v'\n",
						common.Definitions.Pixa.Emoji,
						common.Definitions.Pixa.AppName,
						options,
						strings.Join(args, "/"),
					)
					// ---> execute application core with the parameter set (native)
					//
					// appErr = runApplication(native)
					//
				} else {
					return xv
				}
			} else {
				return err
			}

			return appErr
		},
	}

	paramSet := assistant.NewParamSet[MagickParameterSet](magickCommand)

	container.MustRegisterRootedCommand(magickCommand)
	container.MustRegisterParamSet(MagickPsName, paramSet)

	return magickCommand
}
