package command

import (
	"fmt"
	"os"

	"github.com/cubiest/jibberjabber"
	"github.com/samber/lo"
	"github.com/snivilised/cobrass/src/assistant"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/text/language"

	"github.com/snivilised/arcadia/src/i18n"
	xi18n "github.com/snivilised/extendio/i18n"
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
	b.configure()

	// all these string literals here should be made translate-able
	//

	b.container = assistant.NewCobraContainer(
		&cobra.Command{
			Use:     "main",
			Short:   xi18n.Text(i18n.RootCmdShortDescTemplData{}),
			Long:    xi18n.Text(i18n.RootCmdLongDescTemplData{}),
			Version: fmt.Sprintf("'%v'", Version),
			// Uncomment the following line if your bare application
			// has an action associated with it:
			// Run: func(cmd *cobra.Command, args []string) { },
		},
	)

	b.buildRootCommand(b.container)
	buildWidgetCommand(b.container)

	return b.container.Root()
}

type configureOptions struct {
	configFile *string
}

type ConfigureOptionFn func(*configureOptions)

func (b *Bootstrap) configure(options ...ConfigureOptionFn) {
	var configFile string

	o := configureOptions{
		configFile: &configFile,
	}
	for _, fo := range options {
		fo(&o)
	}

	if configFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(configFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".arcadia" (without extension).
		// NB: 'arcadia' should be renamed as appropriate
		//
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		configName := fmt.Sprintf("%v.yml", ApplicationName)
		viper.SetConfigName(configName)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()

	handleLangSetting()

	if err != nil {
		msg := xi18n.Text(i18n.UsingConfigFileTemplData{
			ConfigFileName: viper.ConfigFileUsed(),
		})
		fmt.Fprintln(os.Stderr, msg)
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
				SourceID: xi18n.TranslationSource{Name: ApplicationName},
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
	root := container.Root()
	paramSet := assistant.NewParamSet[RootParameterSet](root)

	paramSet.BindString(&assistant.FlagInfo{
		Name: "config",
		Usage: xi18n.Text(i18n.RootCmdConfigFileUsageTemplData{
			ConfigFileName: ApplicationName,
		}),
		Default:            "",
		AlternativeFlagSet: root.PersistentFlags(),
	}, &paramSet.Native.ConfigFile)

	paramSet.BindValidatedString(&assistant.FlagInfo{
		Name:               "lang",
		Usage:              xi18n.Text(i18n.RootCmdLangUsageTemplData{}),
		Default:            xi18n.DefaultLanguage.Get().String(),
		AlternativeFlagSet: root.PersistentFlags(),
	}, &paramSet.Native.Language, func(value string) error {
		_, err := language.Parse(value)
		return err
	})

	container.MustRegisterParamSet(RootPsName, paramSet)
}
