package command

import (
	"fmt"
	"maps"
	"strings"

	"github.com/snivilised/cobrass"
	"github.com/snivilised/cobrass/src/assistant"
	"github.com/snivilised/cobrass/src/store"
	xi18n "github.com/snivilised/extendio/i18n"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/snivilised/pixa/src/app/proxy"
	"github.com/snivilised/pixa/src/i18n"
	"github.com/snivilised/pixa/src/internal/helpers"
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
	"mode":   "m",
	// families:
	//
	"cpu":        "C", // family: worker-pool
	"now":        "N", // family: worker-pool
	"dry-run":    "D", // family: preview
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
	filesFamName = "files-family"
)

func newShrinkFlagInfoWithShort[T any](usage string, defaultValue T) *assistant.FlagInfo {
	name := strings.Split(usage, " ")[0]
	short := shrinkShortFlags[name]

	return assistant.NewFlagInfo(usage, short, defaultValue)
}

type shrinkParameterSetPtr = *assistant.ParamSet[proxy.ShrinkParameterSet]

func (b *Bootstrap) buildShrinkCommand(container *assistant.CobraContainer) *cobra.Command {
	shrinkCommand := &cobra.Command{
		Use: "shrink",
		Short: i18n.LeadsWith(
			"shrink",
			xi18n.Text(i18n.ShrinkCmdShortDefinitionTemplData{}),
		),
		Long: xi18n.Text(i18n.ShrinkLongDefinitionTemplData{}),

		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("		===> 🌷🌷🌷 Shrink Command...\n")
			var appErr error

			shrinkPS := container.MustGetParamSet(shrinkPsName).(shrinkParameterSetPtr) //nolint:errcheck // is Must call

			if validationErr := shrinkPS.Validate(); validationErr == nil {
				// optionally invoke cross field validation
				//
				if xvErr := shrinkPS.CrossValidate(func(ps *proxy.ShrinkParameterSet) error {
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

					fmt.Printf("%v %v Running shrink, with options: '%v', args: '%v'\n",
						AppEmoji, ApplicationName, changed, strings.Join(args, "/"),
					)

					inputs := b.getShrinkInputs()
					inputs.Root.ParamSet.Native.Directory = helpers.ResolvePath(args[0])

					// Apply fallbacks, ie user didn't specify flag on command line
					// so fallback to one defined in config. This is supposed to
					// work transparently with Viper, but this doesn't work with
					// custom locations; ie no-files is defined under sampler, but
					// viper would expect to see it at the root. Even so, still found
					// that viper would fail to pick up this value, so implementing
					// the fall back manually here.
					//
					if inputs.Root.ParamSet.Native.IsSampling {
						if !flagSet.Changed("no-files") && b.SamplerCFG.NoFiles() > 0 {
							inputs.Root.ParamSet.Native.NoFiles = b.SamplerCFG.NoFiles()
						}
					}

					appErr = proxy.EnterShrink(
						inputs,
						b.OptionsInfo.Program,
						b.OptionsInfo.Config.Viper,
						b.ProfilesCFG,
						b.SchemesCFG,
						b.SamplerCFG,
						b.AdvancedCFG,
						b.Vfs,
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

	paramSet := assistant.NewParamSet[proxy.ShrinkParameterSet](shrinkCommand)

	// --output(o)
	//
	const (
		defaultOutputPath = ""
	)

	paramSet.BindValidatedString(
		newShrinkFlagInfoWithShort(
			xi18n.Text(i18n.ShrinkCmdOutputPathParamUsageTemplData{}),
			defaultOutputPath,
		),
		&paramSet.Native.OutputPath, func(s string, f *pflag.Flag) error {
			if f.Changed && !b.Vfs.DirectoryExists(s) {
				return i18n.NewOutputPathDoesNotExistError(s)
			}

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
			xi18n.Text(i18n.ShrinkCmdTrashPathParamUsageTemplData{}),
			defaultTrashPath,
		),
		&paramSet.Native.TrashPath, func(s string, f *pflag.Flag) error {
			if f.Changed && !b.Vfs.DirectoryExists(s) {
				return i18n.NewOutputPathDoesNotExistError(s)
			}

			return nil
		},
	)

	// --mode(m)
	//
	const (
		defaultMode = "preserve"
	)

	paramSet.Native.ModeEn = proxy.ModeEnumInfo.NewValue()

	paramSet.BindValidatedEnum(
		newShrinkFlagInfoWithShort(
			xi18n.Text(i18n.ShrinkCmdModeParamUsageTemplData{}),
			defaultMode,
		),
		&paramSet.Native.ModeEn.Source,
		func(value string, f *pflag.Flag) error {
			if f.Changed && !(proxy.ModeEnumInfo.IsValid(value)) {
				acceptableSet := proxy.ModeEnumInfo.AcceptablePrimes()

				return i18n.NewModeError(value, acceptableSet)
			}

			return nil
		},
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
			xi18n.Text(i18n.ShrinkCmdGaussianBlurParamUsageTemplData{}),
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

	paramSet.Native.ThirdPartySet.SamplingFactorEn = proxy.SamplingFactorEnumInfo.NewValue()

	paramSet.BindValidatedEnum(
		newShrinkFlagInfoWithShort(
			xi18n.Text(i18n.ShrinkCmdSamplingFactorParamUsageTemplData{}),
			defaultSamplingFactor,
		),
		&paramSet.Native.ThirdPartySet.SamplingFactorEn.Source,
		func(value string, f *pflag.Flag) error {
			if f.Changed && !(proxy.SamplingFactorEnumInfo.IsValid(value)) {
				acceptableSet := proxy.SamplingFactorEnumInfo.AcceptablePrimes()

				return i18n.NewInvalidSamplingFactorError(value, acceptableSet)
			}
			return nil
		},
	)

	// --interlace(i)
	//
	const (
		defaultInterlace = "plane"
	)

	paramSet.Native.ThirdPartySet.InterlaceEn = proxy.InterlaceEnumInfo.NewValue()

	paramSet.BindValidatedEnum(
		newShrinkFlagInfoWithShort(
			xi18n.Text(i18n.ShrinkCmdInterlaceParamUsageTemplData{}),
			defaultInterlace,
		),
		&paramSet.Native.ThirdPartySet.InterlaceEn.Source,
		func(value string, f *pflag.Flag) error {
			if f.Changed && !(proxy.InterlaceEnumInfo.IsValid(value)) {
				acceptableSet := proxy.InterlaceEnumInfo.AcceptablePrimes()

				return i18n.NewInterlaceError(value, acceptableSet)
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
			xi18n.Text(i18n.ShrinkCmdStripParamUsageTemplData{}),
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
			xi18n.Text(i18n.ShrinkCmdQualityParamUsageTemplData{}),
			defaultQuality,
		),
		&paramSet.Native.ThirdPartySet.Quality,
		minQuality,
		maxQuality,
	)

	// family: files [--files-gb(G), --files-rx(X)]
	//
	// The files filter family is not required on the root, so we define it
	// here instead.
	//
	filesFam := assistant.NewParamSet[store.FilesFilterParameterSet](shrinkCommand)
	filesFam.Native.BindAll(filesFam)

	paramSet.Native.KnownBy = thirdPartyFlags

	b.viper()

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
	container.MustRegisterParamSet(filesFamName, filesFam)

	// TODO: we might need to code this via an anonymous func, store the vfs on
	// the bootstrap, then access it from the func, instead of using
	// validatePositionalArgs
	//
	// shrinkCommand.Args = validatePositionalArgs

	return shrinkCommand
}

func (b *Bootstrap) getShrinkInputs() *proxy.ShrinkCommandInputs {
	return &proxy.ShrinkCommandInputs{
		Root: b.getRootInputs(),
		ParamSet: b.Container.MustGetParamSet(
			shrinkPsName,
		).(*assistant.ParamSet[proxy.ShrinkParameterSet]),
		FilesFam: b.Container.MustGetParamSet(
			filesFamName,
		).(*assistant.ParamSet[store.FilesFilterParameterSet]),
	}
}
