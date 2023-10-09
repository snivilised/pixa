package command

import (
	"fmt"
	"os"
	"regexp"

	"github.com/cubiest/jibberjabber"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/text/language"

	"github.com/snivilised/cobrass/src/assistant"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	ci18n "github.com/snivilised/cobrass/src/assistant/i18n"
	xi18n "github.com/snivilised/extendio/i18n"
	"github.com/snivilised/extendio/xfs/utils"
	"github.com/snivilised/pixa/src/app/magick"
	"github.com/snivilised/pixa/src/i18n"
)

type LocaleDetector interface {
	Scan() language.Tag
}

// Jabber is a LocaleDetector implemented using jibberjabber.
type Jabber struct {
}

// Scan returns the detected language tag.
func (j *Jabber) Scan() language.Tag {
	lang, _ := jibberjabber.DetectIETF()
	return language.MustParse(lang)
}

func validatePositionalArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(1)(cmd, args); err != nil {
		return err
	}

	directory := magick.ResolvePath(args[0])

	if !utils.Exists(directory) {
		return xi18n.NewPathNotFoundError("shrink directory", directory)
	}

	return nil
}

type ConfigInfo struct {
	Name       string
	ConfigType string
	ConfigPath string
	Viper      configuration.ViperConfig
}

// Bootstrap represents construct that performs start up of the cli
// without resorting to the use of Go's init() mechanism and minimal
// use of package global variables.
type Bootstrap struct {
	Container *assistant.CobraContainer
	options   ConfigureOptions
}

type ConfigureOptions struct {
	Detector LocaleDetector
	Executor magick.Executor
	Config   ConfigInfo
}

type ConfigureOptionFn func(*ConfigureOptions)

// Root builds the command tree and returns the root command, ready
// to be executed.
func (b *Bootstrap) Root(options ...ConfigureOptionFn) *cobra.Command {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	b.options = ConfigureOptions{
		Detector: &Jabber{},
		Executor: &ProgramExecutor{
			Name: "magick",
		},
		Config: ConfigInfo{
			Name:       ApplicationName,
			ConfigType: "yaml",
			ConfigPath: home,
			Viper:      &configuration.GlobalViperConfig{},
		},
	}

	if _, err := b.options.Executor.Look(); err != nil {
		b.options.Executor = &DummyExecutor{
			Name: b.options.Executor.ProgName(),
		}
	}

	for _, fo := range options {
		fo(&b.options)
	}

	b.configure()

	// JUST TEMPORARY: make the executor the dummy for safety
	//
	b.options.Executor = &DummyExecutor{
		Name: "magick",
	}

	fmt.Printf("===> ⚠️⚠️⚠️ USING DUMMY EXECUTOR !!!!\n")

	b.Container = assistant.NewCobraContainer(
		&cobra.Command{
			Use:     "main",
			Short:   xi18n.Text(i18n.RootCmdShortDescTemplData{}),
			Long:    xi18n.Text(i18n.RootCmdLongDescTemplData{}),
			Version: fmt.Sprintf("'%v'", Version),
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Printf("		===> 🌷🌷🌷 Root Command...\n")

				rps := b.Container.MustGetParamSet(RootPsName).(magick.RootParameterSetPtr) //nolint:errcheck // is Must call
				rps.Native.Directory = magick.ResolvePath(args[0])

				if rps.Native.CPU {
					rps.Native.NoW = 0
				}

				// ---> execute root core
				//
				return magick.EnterRoot(rps, b.options.Executor, b.options.Config.Viper)
			},
		},
	)

	b.buildRootCommand(b.Container)
	b.buildMagickCommand(b.Container)
	b.buildShrinkCommand(b.Container)

	return b.Container.Root()
}

func (b *Bootstrap) configure() {
	vc := b.options.Config.Viper
	ci := b.options.Config

	vc.SetConfigName(ci.Name)
	vc.SetConfigType(ci.ConfigType)
	vc.AddConfigPath(ci.ConfigPath)
	vc.AutomaticEnv()

	err := vc.ReadInConfig()

	handleLangSetting(vc)

	if err != nil {
		msg := xi18n.Text(i18n.UsingConfigFileTemplData{
			ConfigFileName: vc.ConfigFileUsed(),
		})

		fmt.Fprintln(os.Stderr, msg)

		fmt.Printf("💥 error reading config path: '%v' \n", err)
	}
}

func handleLangSetting(config configuration.ViperConfig) {
	tag := lo.TernaryF(config.InConfig("lang"),
		func() language.Tag {
			lang := viper.GetString("lang")
			parsedTag, err := language.Parse(lang)

			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			return parsedTag
		},
		func() language.Tag {
			return xi18n.DefaultLanguage.Get()
		},
	)

	err := xi18n.Use(func(uo *xi18n.UseOptions) {
		uo.Tag = tag
		uo.From = xi18n.LoadFrom{
			Sources: xi18n.TranslationFiles{
				SourceID: xi18n.TranslationSource{
					Name: ApplicationName,
				},
				ci18n.CobrassSourceID: xi18n.TranslationSource{
					Name: "cobrass",
				},
			},
		}
	})

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (b *Bootstrap) buildRootCommand(container *assistant.CobraContainer) {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	//
	rootCommand := container.Root()
	paramSet := assistant.NewParamSet[magick.RootParameterSet](rootCommand)

	// --cpu(C)
	//
	paramSet.BindBool(&assistant.FlagInfo{
		Name:               "cpu",
		Usage:              "cpu sets the number of workers to the number of available processors",
		Short:              "C",
		Default:            false,
		AlternativeFlagSet: rootCommand.PersistentFlags(),
	}, &paramSet.Native.CPU)

	// --dry-run(D)
	//
	paramSet.BindBool(&assistant.FlagInfo{
		Name:               "dry-run",
		Usage:              "dry-run shrink op",
		Short:              "D",
		Default:            false,
		AlternativeFlagSet: rootCommand.PersistentFlags(),
	}, &paramSet.Native.DryRun)

	// --files-gb(G)
	//
	paramSet.BindString(&assistant.FlagInfo{
		Name: "files-gb",
		Usage: i18n.LeadsWith(
			"files-gb",
			xi18n.Text(i18n.RootCmdFilesGlobParamUsageTemplData{}),
		),
		Short:              "G",
		Default:            "",
		AlternativeFlagSet: rootCommand.PersistentFlags(),
	}, &paramSet.Native.FilesGlob)

	// --files-rx(X)
	//
	paramSet.BindValidatedString(&assistant.FlagInfo{
		Name: "files-rx",
		Usage: i18n.LeadsWith(
			"files-rx",
			xi18n.Text(i18n.RootCmdFilesRegExParamUsageTemplData{}),
		),
		Short:              "X",
		Default:            "",
		AlternativeFlagSet: rootCommand.PersistentFlags(),
	}, &paramSet.Native.FilesRexEx, func(value string, _ *pflag.Flag) error {
		_, err := regexp.Compile(value)
		return err
	})

	// --folder-gb(z)
	//
	paramSet.BindString(&assistant.FlagInfo{
		Name: "folder-gb",
		Usage: i18n.LeadsWith(
			"folder-gb",
			xi18n.Text(i18n.RootCmdFolderGlobParamUsageTemplData{}),
		),
		Short:              "z",
		Default:            "",
		AlternativeFlagSet: rootCommand.PersistentFlags(),
	}, &paramSet.Native.FolderGlob)

	// --folder-rx(y)
	//
	paramSet.BindValidatedString(&assistant.FlagInfo{
		Name: "folder-rx",
		Usage: i18n.LeadsWith(
			"folder-rx",
			xi18n.Text(i18n.RootCmdFolderRexExParamUsageTemplData{}),
		),
		Short:              "y",
		Default:            "",
		AlternativeFlagSet: rootCommand.PersistentFlags(),
	}, &paramSet.Native.FolderRexEx, func(value string, _ *pflag.Flag) error {
		_, err := regexp.Compile(value)
		return err
	})

	// --lang
	//
	paramSet.BindValidatedString(&assistant.FlagInfo{
		Name: "lang",
		Usage: i18n.LeadsWith(
			"lang",
			xi18n.Text(i18n.RootCmdLangUsageTemplData{}),
		),
		Default:            xi18n.DefaultLanguage.Get().String(),
		AlternativeFlagSet: rootCommand.PersistentFlags(),
	}, &paramSet.Native.Language, func(value string, _ *pflag.Flag) error {
		_, err := language.Parse(value)
		return err
	})

	// --now(N)
	//
	const (
		defaultNoW = -1
	)

	paramSet.BindInt(&assistant.FlagInfo{
		Name:               "now",
		Usage:              "now represents number of workers in pool",
		Short:              "N",
		Default:            defaultNoW,
		AlternativeFlagSet: rootCommand.PersistentFlags(),
	}, &paramSet.Native.NoW)

	// --profile(p)
	//
	paramSet.BindString(&assistant.FlagInfo{
		Name:               "profile",
		Usage:              "profile identifies a named group of flag options loaded from config",
		Short:              "P",
		Default:            "",
		AlternativeFlagSet: rootCommand.PersistentFlags(),
	}, &paramSet.Native.Profile)

	// parameter groups
	//
	rootCommand.MarkFlagsMutuallyExclusive("files-rx", "files-gb")
	rootCommand.MarkFlagsMutuallyExclusive("folder-rx", "folder-gb")
	rootCommand.MarkFlagsMutuallyExclusive("now", "cpu")

	rootCommand.Args = validatePositionalArgs

	container.MustRegisterParamSet(RootPsName, paramSet)
}
