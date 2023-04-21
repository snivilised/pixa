package command

import (
	"fmt"
	"strings"

	"github.com/snivilised/cobrass/src/assistant"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CLIENT-TODO: remove this dummy command and replace with application/library
// relevant alternative(s)

type MagickParameterSet struct {
	// remove the follow:
	//
	Directory string
	Concise   bool
	Pattern   string
	Threshold uint
}

const MagickPsName = "magick-ps"

func buildMagickCommand(container *assistant.CobraContainer) *cobra.Command {
	// to test: pixa magick -d ./some-existing-file -p "P?<date>" -t 30
	//
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
						AppEmoji, ApplicationName, options, strings.Join(args, "/"),
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

	// If you want to disable the magick command but keep it in the project for reference
	// purposes, then simply comment out the following 2 register calls:
	// (Warning, this may just create dead code and result in lint failure so tread
	// carefully.)
	//
	container.MustRegisterRootedCommand(magickCommand)
	container.MustRegisterParamSet(MagickPsName, paramSet)

	return magickCommand
}
