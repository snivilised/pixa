package cfg

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/samber/lo"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	ci18n "github.com/snivilised/cobrass/src/assistant/i18n"
	xi18n "github.com/snivilised/extendio/i18n"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
)

func New(vc configuration.ViperConfig,
	ci *common.ConfigInfo,
	sourceID string,
	applicationName string,
) common.ConfigRunner {
	return &configRunner{
		vc:              vc,
		ci:              ci,
		sourceID:        sourceID,
		applicationName: applicationName,
	}
}

type configRunner struct {
	vc              configuration.ViperConfig
	ci              *common.ConfigInfo
	sourceID        string
	applicationName string
}

func (c configRunner) Run() error {
	c.vc.SetConfigName(c.ci.Name)
	c.vc.SetConfigType(c.ci.ConfigType)
	c.vc.AddConfigPath(c.path())
	c.vc.AutomaticEnv()

	err := c.vc.ReadInConfig()

	c.handleLangSetting(c.vc)

	return err
}

func (c configRunner) path() string {
	configPath := c.ci.ConfigPath

	if configPath == "" {
		configPath, _ = c.vc.Get("PIXA-HOME").(string)

		fmt.Printf("---> ✨ PIXA-HOME found in environment: '%v'\n", configPath)
	}

	if configPath == "" {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		defaultPath := filepath.Join("snivilised", "pixa")
		configPath = filepath.Join(home, defaultPath)
	}

	return configPath
}

func (c configRunner) handleLangSetting(config configuration.ViperConfig) {
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
				c.sourceID: xi18n.TranslationSource{
					Name: c.applicationName,
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
