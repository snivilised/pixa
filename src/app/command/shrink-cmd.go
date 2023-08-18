package command

import (
	"fmt"
	"strings"

	"github.com/snivilised/cobrass/src/assistant"
	xi18n "github.com/snivilised/extendio/i18n"
	"github.com/snivilised/extendio/xfs/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/snivilised/pixa/src/app/magick"
	"github.com/snivilised/pixa/src/i18n"
)

// We define all the options here, even the ones inherited from the root
// command, because doing so allows us to see the whole set of options
// applicable to the shrink command in a single place and aid the assignment
// of the short flag names.
//
// NB: we use files instead of file, because these filters are compound
// and we use capitals for the short forms of files filters, to denote
// compound filter. If files filter was not compound, it would be named
// file and the short forms would be x and g instead of X and G.
var shrinkShortFlags = map[string]string{
	// shrink specific:
	//
	"mirror-path": "r",
	"mode":        "m",
	// core:
	//
	"gaussian-blur":   "b",
	"sampling-factor": "f",
	"interlace":       "i",
	"strip":           "s",
	"quality":         "q",
	// root:
	//
	"preview":     "P",
	"folder-rx":   "y",
	"folder-glob": "z",
	"files-rx":    "X",
	"files-glob":  "G",
}

const shrinkPsName = "shrink-ps"

func newShrinkFlagInfo[T any](usage string, defaultValue T) *assistant.FlagInfo {
	name := strings.Split(usage, " ")[0]
	short := shrinkShortFlags[name]

	return assistant.NewFlagInfo(usage, short, defaultValue)
}

type shrinkParameterSetPtr = *assistant.ParamSet[magick.ShrinkParameterSet]

func buildShrinkCommand(container *assistant.CobraContainer) *cobra.Command {
	shrinkCommand := &cobra.Command{
		Use: "shrink",
		Short: i18n.LeadsWith(
			"shrink",
			xi18n.Text(i18n.ShrinkCmdShortDefinitionTemplData{}),
		),
		Long: xi18n.Text(i18n.ShrinkLongDefinitionTemplData{}),

		RunE: func(cmd *cobra.Command, args []string) error {
			var appErr error

			ps := container.MustGetParamSet(shrinkPsName).(shrinkParameterSetPtr) //nolint:errcheck // is Must call

			if validationErr := ps.Validate(); validationErr == nil {
				// optionally invoke cross field validation
				//
				if xvErr := ps.CrossValidate(func(ps *magick.ShrinkParameterSet) error {
					// cross validation not currently required
					//
					return nil
				}); xvErr == nil {
					options := []string{}
					cmd.Flags().Visit(func(f *pflag.Flag) {
						options = append(options, fmt.Sprintf("--%v=%v", f.Name, f.Value))
					})

					fmt.Printf("%v %v Running shrink, with options: '%v', args: '%v'\n",
						AppEmoji, ApplicationName, options, strings.Join(args, "/"),
					)

					if cmd.Flags().Changed("gaussian-blur") {
						fmt.Printf("💠 Blur defined with value: '%v'\n", cmd.Flag("gaussian-blur").Value)
					}

					if cmd.Flags().Changed("sampling-factor") {
						fmt.Printf("💠 Blur defined with value: '%v'\n", cmd.Flag("sampling-factor").Value)
					}

					// Get inherited parameters
					//
					rps := container.MustGetParamSet(RootPsName).(magick.RootParameterSetPtr) //nolint:errcheck // is Must call
					rps.Native.Directory = magick.ResolvePath(args[0])

					// ---> execute application core with the parameter set (native)
					//
					appErr = magick.EnterShrink(rps, ps)
				} else {
					return xvErr
				}
			} else {
				return validationErr
			}

			return appErr
		},
	}

	paramSet := assistant.NewParamSet[magick.ShrinkParameterSet](shrinkCommand)

	// Gaussian Blur
	//
	const (
		defaultBlur = float32(0.05)
		minBlur     = float32(0.01)
		maxBlur     = float32(1.0)
	)

	paramSet.BindValidatedFloat32Within(
		newShrinkFlagInfo(
			xi18n.Text(i18n.ShrinkCmdGaussianBlurParamUsageTemplData{}),
			defaultBlur,
		),
		&paramSet.Native.Gaussian,
		minBlur,
		maxBlur,
	)

	// Sampling Factor
	//
	paramSet.Native.FactorEn = magick.SamplingFactorEnumInfo.NewValue()

	paramSet.BindValidatedEnum(
		newShrinkFlagInfo(
			xi18n.Text(i18n.ShrinkCmdSamplingFactorParamUsageTemplData{}),
			"4:2:0",
		),
		&paramSet.Native.FactorEn.Source,
		func(value string, f *pflag.Flag) error {
			if f.Changed && !(magick.SamplingFactorEnumInfo.IsValid(value)) {
				acceptableSet := magick.SamplingFactorEnumInfo.AcceptablePrimes()

				return i18n.NewInvalidSamplingFactorError(value, acceptableSet)
			}
			return nil
		},
	)

	// Interlace
	//
	paramSet.Native.InterlaceEn = magick.InterlaceEnumInfo.NewValue()

	paramSet.BindValidatedEnum(
		newShrinkFlagInfo(
			xi18n.Text(i18n.ShrinkCmdInterlaceParamUsageTemplData{}),
			"plane",
		),
		&paramSet.Native.InterlaceEn.Source,
		func(value string, f *pflag.Flag) error {
			if f.Changed && !(magick.InterlaceEnumInfo.IsValid(value)) {
				acceptableSet := magick.InterlaceEnumInfo.AcceptablePrimes()

				return i18n.NewInterlaceError(value, acceptableSet)
			}

			return nil
		},
	)

	// Strip
	//
	paramSet.BindBool(
		newShrinkFlagInfo(
			xi18n.Text(i18n.ShrinkCmdStripParamUsageTemplData{}),
			false,
		),
		&paramSet.Native.Strip,
	)

	// Quality
	//
	const (
		defaultQuality = int(80)
		minQuality     = int(0)
		maxQuality     = int(100)
	)

	paramSet.BindValidatedIntWithin(
		newShrinkFlagInfo(
			xi18n.Text(i18n.ShrinkCmdQualityParamUsageTemplData{}),
			defaultQuality,
		),
		&paramSet.Native.Quality,
		minQuality,
		maxQuality,
	)

	// Mirror Path
	//
	paramSet.BindValidatedString(
		newShrinkFlagInfo(
			xi18n.Text(i18n.ShrinkCmdMirrorPathParamUsageTemplData{}),
			"",
		),
		&paramSet.Native.MirrorPath, func(s string, f *pflag.Flag) error {
			if f.Changed && !utils.FolderExists(s) {
				return i18n.NewMirrorPathDoesNotExistError(s)
			}

			return nil
		},
	)

	// Mode
	//
	paramSet.Native.ModeEn = magick.ModeEnumInfo.NewValue()

	paramSet.BindValidatedEnum(
		newShrinkFlagInfo(
			xi18n.Text(i18n.ShrinkCmdModeParamUsageTemplData{}),
			"preserve",
		),
		&paramSet.Native.ModeEn.Source,
		func(value string, f *pflag.Flag) error {
			if f.Changed && !(magick.ModeEnumInfo.IsValid(value)) {
				acceptableSet := magick.ModeEnumInfo.AcceptablePrimes()

				return i18n.NewModeError(value, acceptableSet)
			}

			return nil
		},
	)

	// 📌A note about cobra args validation: cmd.ValidArgs lets you define
	// a list of all allowable tokens for positional args. Just define
	// ValidArgs, eg:
	// shrinkCommand.ValidArgs = []string{"foo", "bar", "baz"}
	// and then set the args validation function cmd.Args to cobra.OnlyValidArgs
	// ie:
	// shrinkCommand.Args = cobra.OnlyValidArgs
	// With this in place, the user can only type positional args which are in
	// the set defined, ie {"foo", "bar", "baz"}.
	//
	// Since the shrink command only needs a single 'directory' positional arg,
	// all we need is to set the exact no of args to 1. We don;t need to define
	// ValidArgs since there is no closed set directories we can define. ValidArgs
	// is suitable when all positional args can behave like an enum, where there
	// is a finite set of valid values.
	//
	container.MustRegisterRootedCommand(shrinkCommand)
	container.MustRegisterParamSet(shrinkPsName, paramSet)

	shrinkCommand.Args = validatePositionalArgs

	return shrinkCommand
}
