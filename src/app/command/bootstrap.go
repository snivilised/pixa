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

// Bootstrap represents construct that performs start up of the cli
// without resorting to the use of Go's init() mechanism and minimal
// use of package global variables.
type Bootstrap struct {
	Detector  LocaleDetector
	container *assistant.CobraContainer
}

// Root builds the command tree and returns the root command, ready
// to be executed.
func (b *Bootstrap) Root() *cobra.Command {
	b.configure(func(co *configureOptions) {
		// ---> co.configFile = "~/pixa.yml"
	})

	b.container = assistant.NewCobraContainer(
		&cobra.Command{
			Use:     "main",
			Short:   xi18n.Text(i18n.RootCmdShortDescTemplData{}),
			Long:    xi18n.Text(i18n.RootCmdLongDescTemplData{}),
			Version: fmt.Sprintf("'%v'", Version),
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Printf("		===> 🌷🌷🌷 Root Command...\n")

				rps := b.container.MustGetParamSet(RootPsName).(magick.RootParameterSetPtr) //nolint:errcheck // is Must call
				rps.Native.Directory = magick.ResolvePath(args[0])

				// ---> execute root core
				//
				return magick.EnterRoot(rps)
			},
		},
	)

	b.buildRootCommand(b.container)
	buildMagickCommand(b.container)
	buildShrinkCommand(b.container)

	return b.container.Root()
}

type configureOptions struct {
	configFile string
}

type ConfigureOptionFn func(*configureOptions)

func (b *Bootstrap) configure(options ...ConfigureOptionFn) {
	var configFile string

	o := configureOptions{
		configFile: configFile,
	}
	for _, fo := range options {
		fo(&o)
	}

	if configFile != "" {
		fmt.Printf("💛 setting explicit config path; '%v'\n", configFile)
		// Use config file from the flag.
		viper.SetConfigFile(configFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		fmt.Printf("💚 (nil-config-file) using default config path \n")

		// Search config in home directory with name ".pixa" (without extension).
		// NB: 'arcadia' should be renamed as appropriate
		//
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(ApplicationName)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()

	handleLangSetting()

	msg := xi18n.Text(i18n.UsingConfigFileTemplData{
		ConfigFileName: viper.ConfigFileUsed(),
	})

	if err != nil {
		fmt.Fprintln(os.Stderr, msg)

		fmt.Printf("💥 error reading config path: '%v' \n", err)
	} else {
		fmt.Printf("🧡 '%v' \n", msg)

		gb := viper.GetString("gaussian-blur")
		if gb != "" {
			fmt.Printf("--> 💝 found blur in config: '%v' \n", gb)
		}
	}
}

func handleLangSetting() {
	tag := lo.TernaryF(viper.InConfig("lang"),
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

	paramSet.BindBool(&assistant.FlagInfo{
		Name:               "viper",
		Usage:              "viper defines whether to use viper configuration",
		Default:            true,
		AlternativeFlagSet: rootCommand.PersistentFlags(),
	}, &paramSet.Native.Viper)

	paramSet.BindString(&assistant.FlagInfo{
		Name: "config",
		Usage: i18n.LeadsWith(
			"config",
			xi18n.Text(i18n.RootCmdConfigFileUsageTemplData{}),
		),
		Default:            "pixa.yml",
		AlternativeFlagSet: rootCommand.PersistentFlags(),
	}, &paramSet.Native.ConfigFile)

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

	// FolderRexEx
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

	// FolderGlob
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

	// FilesRexEx
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

	// FilesGlob
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

	// Preview
	//
	paramSet.BindBool(&assistant.FlagInfo{
		Name:               "preview",
		Usage:              "preview shrink op",
		Short:              "P",
		Default:            false,
		AlternativeFlagSet: rootCommand.PersistentFlags(),
	}, &paramSet.Native.Preview)

	rootCommand.MarkFlagsMutuallyExclusive("files-rx", "files-gb")
	rootCommand.MarkFlagsMutuallyExclusive("folder-rx", "folder-gb")

	rootCommand.Args = validatePositionalArgs

	container.MustRegisterParamSet(RootPsName, paramSet)
}
