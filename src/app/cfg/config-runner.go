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

type ConfigRunner struct {
	ViperConfig     configuration.ViperConfig
	ConfigInfo      *common.ConfigInfo
	SourceID        string
	ApplicationName string
}

func (c ConfigRunner) Run() error {
	c.ViperConfig.SetConfigName(c.ConfigInfo.Name)
	c.ViperConfig.SetConfigType(c.ConfigInfo.ConfigType)
	c.ViperConfig.AddConfigPath(c.path())
	c.ViperConfig.AutomaticEnv()

	err := c.ViperConfig.ReadInConfig()

	c.handleLangSetting(c.ViperConfig)

	return err
}

func (c ConfigRunner) path() string {
	configPath := c.ConfigInfo.ConfigPath

	if configPath == "" {
		configPath, _ = c.ViperConfig.Get("PIXA-HOME").(string)

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

func (c ConfigRunner) handleLangSetting(config configuration.ViperConfig) {
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
				c.SourceID: xi18n.TranslationSource{
					Name: c.ApplicationName,
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
