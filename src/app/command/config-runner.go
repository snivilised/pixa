package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/snivilised/cobrass/src/assistant/configuration"
	xi18n "github.com/snivilised/extendio/i18n"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/snivilised/pixa/src/i18n"
	"github.com/spf13/cobra"
)

type configRunner struct {
	vc configuration.ViperConfig
	ci *common.ConfigInfo
}

func (c configRunner) Run() error {
	c.vc.SetConfigName(c.ci.Name)
	c.vc.SetConfigType(c.ci.ConfigType)
	c.vc.AddConfigPath(c.path())
	c.vc.AutomaticEnv()

	err := c.vc.ReadInConfig()

	handleLangSetting(c.vc)

	if err != nil {
		msg := xi18n.Text(i18n.UsingConfigFileTemplData{
			ConfigFileName: c.vc.ConfigFileUsed(),
		})

		fmt.Fprintln(os.Stderr, msg)

		fmt.Printf("💥 error reading config path: '%v' \n", err)
	}

	return nil
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
