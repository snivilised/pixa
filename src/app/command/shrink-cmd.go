package command

import (
	"fmt"
	"log/slog"
	"maps"
	"strings"

	"github.com/snivilised/cobrass"
	"github.com/snivilised/cobrass/src/assistant"
	"github.com/snivilised/cobrass/src/store"
	"github.com/snivilised/extendio/xfs/utils"
	"github.com/snivilised/li18ngo"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/snivilised/pixa/src/app/proxy"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/snivilised/pixa/src/locale"
)

// We define all the options here, even the ones inherited from the root
// command, because doing so allows us to see the whole set of options
// applicable to the shrink command in a single place and aid the assignment
// of the short flag names. This also helps when using families, as it
// reminds us not to attempt to reuse a short flag that has already been
// allocated.
//
// NB: we use files instead of file, because these filters are compound
// and we use capitals for the short forms of files filters, to denote
// compound filter. If files filter was not compound, it would be named
// file and the short forms would be x and g instead of X and G.

var thirdPartyFlags = cobrass.KnownByCollection{
	// third-party: (perhaps third party parameters should not have short codes)
	//
	"gaussian-blur":   "b",
	"sampling-factor": "f",
	"interlace":       "i",
	"strip":           "s",
	"quality":         "q",
}

var shrinkShortFlags = cobrass.KnownByCollection{
	// shrink specific:
	//
	"output": "o",
	"trash":  "t",
	"cuddle": "c",
	// families:
	//
	"cpu":        "C", // family: worker-pool
	"now":        "N", // family: worker-pool
	"dry-run":    "D", // family: preview
	"files":      "F", // family: filter
	"files-gb":   "G", // family: filter
	"files-rx":   "X", // family: filter
	"folders-gb": "Z", // family: filter
	"folders-rx": "Y", // family: filter
	"profile":    "P", // family: profile
	"scheme":     "S", // family: profile
}

func init() {
	maps.Copy(shrinkShortFlags, thirdPartyFlags)
}

const (
	shrinkPsName = "shrink-ps"
	polyFamName  = "poly-family"
)

func newShrinkFlagInfoWithShort[T any](usage string, defaultValue T) *assistant.FlagInfo {
	name := strings.Split(usage, " ")[0]
	short := shrinkShortFlags[name]

	return assistant.NewFlagInfo(usage, short, defaultValue)
}

type shrinkParameterSetPtr = *assistant.ParamSet[common.ShrinkParameterSet]

func (b *Bootstrap) buildShrinkCommand(container *assistant.CobraContainer) *cobra.Command {
	shrinkCommand := &cobra.Command{
		Use: "shrink",
		Short: locale.LeadsWith(
			"shrink",
			li18ngo.Text(locale.ShrinkCmdShortDefinitionTemplData{}),
		),
		Long: li18ngo.Text(locale.ShrinkLongDefinitionTemplData{}),

		RunE: func(cmd *cobra.Command, args []string) error {
			var appErr error

			shrinkPS := container.MustGetParamSet(shrinkPsName).(shrinkParameterSetPtr) //nolint:errcheck // is Must call

			if validationErr := shrinkPS.Validate(); validationErr == nil {
				// optionally invoke cross field validation
				//
				if xvErr := shrinkPS.CrossValidate(func(_ *common.ShrinkParameterSet) error {
					// cross validation not currently required
					//
					return nil
				}); xvErr == nil {
					flagSet := cmd.Flags()
					changed := assistant.GetThirdPartyCL(
						flagSet,
						shrinkPS.Native.ThirdPartySet.KnownBy,
					)

					// changed is incorrect; it only contains the third party args,
					// all the native args are being omitted

					shrinkPS.Native.ThirdPartySet.LongChangedCL = changed

					b.Logger.Info(
						fmt.Sprintf("%v %v running shrink",
							common.Definitions.Pixa.AppName, common.Definitions.Pixa.Emoji,
						),
						slog.String("args", strings.Join(args, "/")),
					)

					inputs := b.getShrinkInputs()

					inputs.Root.ParamSet.Native.Directory = utils.ResolvePath(args[0])

					// Apply fallbacks, ie user didn't specify flag on command line
					// so fallback to one defined in config. This is supposed to
					// work transparently with Viper, but this doesn't work with
					// custom locations; ie no-files is defined under sampler, but
					// viper would expect to see it at the root. Even so, still found
					// that viper would fail to pick up this value, so implementing
					// the fall back manually here.
					//
					if inputs.Root.ParamSet.Native.IsSampling {
						if !flagSet.Changed("no-files") && b.Configs.Sampler.NoFiles() > 0 {
							inputs.Root.ParamSet.Native.NoFiles = b.Configs.Sampler.NoFiles()
						}
					}

					_, appErr = proxy.EnterShrink(
						&proxy.ShrinkParams{
							Inputs:        inputs,
							Viper:         b.OptionsInfo.Config.Viper,
							Logger:        b.Logger,
							Vfs:           b.Vfs,
							Notifications: &b.Notifications,
						},
					)
				} else {
					return xvErr
				}
			} else {
				return validationErr
			}

			return appErr
		},
	}

	paramSet := assistant.NewParamSet[common.ShrinkParameterSet](shrinkCommand)

	// --output(o)
	//
	const (
		defaultOutputPath = ""
	)

	paramSet.BindValidatedString(
		newShrinkFlagInfoWithShort(
			li18ngo.Text(locale.ShrinkCmdOutputPathParamUsageTemplData{}),
			defaultOutputPath,
		),
		&paramSet.Native.OutputPath, func(_ string, _ *pflag.Flag) error {
			// todo: Instead of doing the commented out check, check that the location
			// specified has the correct permission to write
			//
			// if f.Changed && !b.Vfs.DirectoryExists(s) {
			// 	return i18n.NewOutputPathDoesNotExistError(s)
			// }
			return nil
		},
	)

	// --trash(t)
	//
	const (
		defaultTrashPath = ""
	)

	paramSet.BindValidatedString(
		newShrinkFlagInfoWithShort(
			li18ngo.Text(locale.ShrinkCmdTrashPathParamUsageTemplData{}),
			defaultTrashPath,
		),
		&paramSet.Native.TrashPath, func(_ string, _ *pflag.Flag) error {
			// todo: Instead of doing the commented out check, check that the location
			// specified has the correct permission to write
			//
			// if f.Changed && !b.Vfs.DirectoryExists(s) {
			// 	return i18n.NewOutputPathDoesNotExistError(s)
			// }
			return nil
		},
	)

	// --cuddle(c)
	//
	const (
		defaultCuddle = false
	)

	paramSet.BindBool(
		newShrinkFlagInfoWithShort(
			li18ngo.Text(locale.ShrinkCmdCuddleParamUsageTemplData{}),
			defaultCuddle,
		),
		&paramSet.Native.Cuddle,
	)

	// --gaussian-blur(b)
	//
	const (
		defaultBlur = float32(0.05)
		minBlur     = float32(0.01)
		maxBlur     = float32(1.0)
	)

	paramSet.BindValidatedFloat32Within(
		newShrinkFlagInfoWithShort(
			li18ngo.Text(locale.ShrinkCmdGaussianBlurParamUsageTemplData{}),
			defaultBlur,
		),
		&paramSet.Native.ThirdPartySet.GaussianBlur,
		minBlur,
		maxBlur,
	)

	// --sampling-factor(f)
	//
	const (
		defaultSamplingFactor = "4:2:0"
	)

	paramSet.Native.ThirdPartySet.SamplingFactorEn = common.SamplingFactorEnumInfo.NewValue()

	paramSet.BindValidatedEnum(
		newShrinkFlagInfoWithShort(
			li18ngo.Text(locale.ShrinkCmdSamplingFactorParamUsageTemplData{}),
			defaultSamplingFactor,
		),
		&paramSet.Native.ThirdPartySet.SamplingFactorEn.Source,
		func(value string, f *pflag.Flag) error {
			if f.Changed && !(common.SamplingFactorEnumInfo.IsValid(value)) {
				acceptableSet := common.SamplingFactorEnumInfo.AcceptablePrimes()

				return locale.NewInvalidSamplingFactorError(value, acceptableSet)
			}

			return nil
		},
	)

	// --interlace(i)
	//
	const (
		defaultInterlace = "plane"
	)

	paramSet.Native.ThirdPartySet.InterlaceEn = common.InterlaceEnumInfo.NewValue()

	paramSet.BindValidatedEnum(
		newShrinkFlagInfoWithShort(
			li18ngo.Text(locale.ShrinkCmdInterlaceParamUsageTemplData{}),
			defaultInterlace,
		),
		&paramSet.Native.ThirdPartySet.InterlaceEn.Source,
		func(value string, f *pflag.Flag) error {
			if f.Changed && !(common.InterlaceEnumInfo.IsValid(value)) {
				acceptableSet := common.InterlaceEnumInfo.AcceptablePrimes()

				return locale.NewInterlaceError(value, acceptableSet)
			}

			return nil
		},
	)

	// --strip(s)
	//
	const (
		defaultStrip = false
	)

	paramSet.BindBool(
		newShrinkFlagInfoWithShort(
			li18ngo.Text(locale.ShrinkCmdStripParamUsageTemplData{}),
			defaultStrip,
		),
		&paramSet.Native.ThirdPartySet.Strip,
	)

	// --quality(q)
	//
	const (
		defaultQuality = int(80)
		minQuality     = int(0)
		maxQuality     = int(100)
	)

	paramSet.BindValidatedIntWithin(
		newShrinkFlagInfoWithShort(
			li18ngo.Text(locale.ShrinkCmdQualityParamUsageTemplData{}),
			defaultQuality,
		),
		&paramSet.Native.ThirdPartySet.Quality,
		minQuality,
		maxQuality,
	)

	// family: poly [--files(f), --files-rx(X), --folders-gb(Z), --folders-rx(Y)]
	//
	// A file filter is not required on the root, so we define it
	// here instead.
	//
	polyFam := assistant.NewParamSet[store.PolyFilterParameterSet](shrinkCommand)
	polyFam.Native.BindAll(polyFam)

	paramSet.Native.KnownBy = thirdPartyFlags

	// 📌 A note about cobra args validation: cmd.ValidArgs lets you define
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
	container.MustRegisterParamSet(polyFamName, polyFam)

	// TODO: we might need to code this via an anonymous func, store the vfs on
	// the bootstrap, then access it from the func, instead of using
	// validatePositionalArgs
	//
	// shrinkCommand.Args = validatePositionalArgs

	// If we allowed --output to be specified with --cuddle, then that would
	// mean the result files would be written to the output location and then input
	// files would have to follow the results, leaving the origin without
	// the input or the output. This could be seen as excessive and unnecessary.
	// The cuddle option is most useful to the user when running a sample to
	// enable easier comparison of the result with the input. If the user
	// really wants to cuddle, then there should be no need to specify an output.
	// We need to reduce the number of permutations to reduce complexity and
	// the number of required unit tests; particularly for the path-finder.
	// 	The same logic applies to cuddle with trash, except that's its even more
	// acute in this usage scenario, because you never want your new results
	// to be cuddled into the trash location.
	paramSet.Command.MarkFlagsMutuallyExclusive("output", "cuddle")
	paramSet.Command.MarkFlagsMutuallyExclusive("trash", "cuddle")

	return shrinkCommand
}

func (b *Bootstrap) getShrinkInputs() *common.ShrinkCommandInputs {
	return &common.ShrinkCommandInputs{
		Root: b.getRootInputs(),
		ParamSet: b.Container.MustGetParamSet(
			shrinkPsName,
		).(*assistant.ParamSet[common.ShrinkParameterSet]),
		PolyFam: b.Container.MustGetParamSet(
			polyFamName,
		).(*assistant.ParamSet[store.PolyFilterParameterSet]),
	}
}
