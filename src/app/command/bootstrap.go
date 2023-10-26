package command

import (
	"fmt"
	"os"

	"github.com/cubiest/jibberjabber"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/text/language"

	"github.com/snivilised/cobrass/src/assistant"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	ci18n "github.com/snivilised/cobrass/src/assistant/i18n"
	xi18n "github.com/snivilised/extendio/i18n"
	"github.com/snivilised/extendio/xfs/utils"
	"github.com/snivilised/pixa/src/app/proxy"
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

	directory := proxy.ResolvePath(args[0])

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
	Container   *assistant.CobraContainer
	optionsInfo ConfigureOptionsInfo
}

type ConfigureOptionsInfo struct {
	Detector LocaleDetector
	Executor proxy.Executor
	Config   ConfigInfo
}

type ConfigureOptionFn func(*ConfigureOptionsInfo)

// Root builds the command tree and returns the root command, ready
// to be executed.
func (b *Bootstrap) Root(options ...ConfigureOptionFn) *cobra.Command {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	b.optionsInfo = ConfigureOptionsInfo{
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

	if _, err := b.optionsInfo.Executor.Look(); err != nil {
		b.optionsInfo.Executor = &DummyExecutor{
			Name: b.optionsInfo.Executor.ProgName(),
		}
	}

	for _, fo := range options {
		fo(&b.optionsInfo)
	}

	b.configure()

	// JUST TEMPORARY: make the executor the dummy for safety
	//
	b.optionsInfo.Executor = &DummyExecutor{
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

				inputs := b.getRootInputs()
				inputs.ParamSet.Native.Directory = proxy.ResolvePath(args[0])
				if inputs.ParamSet.Native.CPU {
					inputs.ParamSet.Native.NoW = 0
				}

				// ---> execute root core
				//
				return proxy.EnterRoot(inputs, b.optionsInfo.Executor, b.optionsInfo.Config.Viper)
			},
		},
	)

	b.buildRootCommand(b.Container)
	b.buildMagickCommand(b.Container)
	b.buildShrinkCommand(b.Container)

	return b.Container.Root()
}

func (b *Bootstrap) configure() {
	vc := b.optionsInfo.Config.Viper
	ci := b.optionsInfo.Config

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
