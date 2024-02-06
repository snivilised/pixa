package command

import (
	"fmt"
	"log/slog"
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
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/extendio/xfs/utils"
	"github.com/snivilised/pixa/src/app/cfg"
	"github.com/snivilised/pixa/src/app/plog"
	"github.com/snivilised/pixa/src/app/proxy"
	"github.com/snivilised/pixa/src/app/proxy/common"
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
	// TODO: actually, it would be better if we can somehow access the vfs
	// instead of using the util.Exist function
	//
	if err := cobra.ExactArgs(1)(cmd, args); err != nil {
		return err
	}

	directory := utils.ResolvePath(args[0])

	if !utils.Exists(directory) {
		return xi18n.NewPathNotFoundError("shrink directory", directory)
	}

	return nil
}

// Bootstrap represents construct that performs start up of the cli
// without resorting to the use of Go's init() mechanism and minimal
// use of package global variables.
type Bootstrap struct {
	Container   *assistant.CobraContainer
	OptionsInfo ConfigureOptionsInfo
	Configs     *common.Configs
	Vfs         storage.VirtualFS
	Logger      *slog.Logger
}

type ConfigureOptionsInfo struct {
	Detector LocaleDetector
	Config   common.ConfigInfo
	Runner   common.ConfigRunner
}

type ConfigureOptionFn func(*ConfigureOptionsInfo)

// Root builds the command tree and returns the root command, ready
// to be executed.
func (b *Bootstrap) Root(options ...ConfigureOptionFn) *cobra.Command {
	vc := &configuration.GlobalViperConfig{}
	ci := common.ConfigInfo{
		Name:       common.Definitions.Pixa.AppName,
		ConfigType: common.Definitions.Pixa.ConfigType,
		Viper:      vc,
	}

	runner, err := cfg.New(
		&ci,
		common.Definitions.Pixa.SourceID,
		common.Definitions.Pixa.AppName,
		b.Vfs,
	)

	if err != nil {
		// not being able to access the default path is pretty catastrophic,
		// so will terminate immediately if this happens.
		//
		fmt.Println("---> 🔥 can't access home path, terminating.")
		os.Exit(1)
	}

	b.OptionsInfo = ConfigureOptionsInfo{
		Detector: &Jabber{},
		Config:   ci,
		Runner:   runner,
	}

	for _, fo := range options {
		fo(&b.OptionsInfo)
	}

	b.configure()
	b.viper()
	b.Logger = plog.New(b.Configs.Logging, b.Vfs)

	b.Container = assistant.NewCobraContainer(
		&cobra.Command{
			Use:     "main",
			Short:   xi18n.Text(i18n.RootCmdShortDescTemplData{}),
			Long:    xi18n.Text(i18n.RootCmdLongDescTemplData{}),
			Version: fmt.Sprintf("'%v'", Version),
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Printf("		===> 🌷🌷🌷 Root Command...\n")

				inputs := b.getRootInputs()
				inputs.ParamSet.Native.Directory = utils.ResolvePath(args[0])

				if inputs.WorkerPoolFam.Native.CPU {
					inputs.WorkerPoolFam.Native.NoWorkers = 0
				}

				if inputs.ProfileFam.Native.Profile != "" {
					if _, found := b.Configs.Profiles.Profile(inputs.ProfileFam.Native.Profile); !found {
						return fmt.Errorf(
							"no such profile: '%v'", inputs.ProfileFam.Native.Profile,
						)
					}
				}

				if scheme := inputs.ProfileFam.Native.Scheme; scheme != "" {
					if err := b.Configs.Schemes.Validate(scheme, b.Configs.Profiles); err != nil {
						return err
					}
				}

				// ---> execute root core
				//

				return proxy.EnterRoot(
					inputs,
					b.OptionsInfo.Config.Viper,
				)
			},
		},
	)

	b.buildRootCommand(b.Container)
	b.buildMagickCommand(b.Container)
	b.buildShrinkCommand(b.Container)

	return b.Container.Root()
}

func (b *Bootstrap) configure() {
	if err := b.OptionsInfo.Runner.Run(); err != nil {
		msg := xi18n.Text(i18n.UsingConfigFileTemplData{
			ConfigFileName: b.OptionsInfo.Config.Viper.ConfigFileUsed(),
		})

		fmt.Fprintln(os.Stderr, msg)
		fmt.Printf("💥 error reading config path: '%v' \n", err)
		b.exit(err)
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
				common.Definitions.Pixa.SourceID: xi18n.TranslationSource{
					Name: common.Definitions.Pixa.AppName,
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

func (b *Bootstrap) viper() {
	var (
		err error
		m   cfg.MsMasterConfig
	)

	b.Configs, err = m.Read(b.OptionsInfo.Config.Viper)

	// TODO: this needs a refactor to handle errors better, we should be returning errors everywhere
	// an error can occur, not selectively. this method is a perfect bad example.

	if err != nil {
		b.exit(err)
	}
}

func (b *Bootstrap) exit(err error) {
	panic(err)
}
