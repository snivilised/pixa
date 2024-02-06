package cfg

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/samber/lo"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	ci18n "github.com/snivilised/cobrass/src/assistant/i18n"
	"github.com/snivilised/extendio/collections"
	xi18n "github.com/snivilised/extendio/i18n"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
)

var (
	//go:embed default-pixa.yml
	defaultConfig string
)

type (
	tryReadConfigFn func() error
)

func GetDefaultConfigContent() string {
	return defaultConfig
}

func New(
	ci *common.ConfigInfo,
	sourceID string,
	applicationName string,
	vfs storage.VirtualFS,
) (common.ConfigRunner, error) {
	home, err := os.UserHomeDir()

	return &configRunner{
		vc:              ci.Viper,
		ci:              ci,
		sourceID:        sourceID,
		applicationName: applicationName,
		home:            home,
		vfs:             vfs,
	}, err
}

type configRunner struct {
	vc              configuration.ViperConfig
	ci              *common.ConfigInfo
	sourceID        string
	applicationName string
	home            string
	vfs             storage.VirtualFS
}

func (c *configRunner) DefaultPath() string {
	return filepath.Join(c.home, common.Definitions.Pixa.SubPath)
}

func (c *configRunner) Run() error {
	c.vc.SetConfigName(c.ci.Name)
	c.vc.SetConfigType(c.ci.ConfigType)
	c.vc.AutomaticEnv()
	c.vc.AddConfigPath(c.path())

	err := c.read()

	c.handleLangSetting(c.vc)

	return err
}

func (c *configRunner) path() string {
	configPath := c.ci.ConfigPath

	if configPath == "" {
		configPath, _ = c.vc.Get(common.Definitions.Environment.Home).(string)
	}

	if configPath == "" {
		configPath = c.DefaultPath()
	} else {
		fmt.Printf("---> ✨ PIXA_HOME found in environment: '%v'\n", configPath)
	}

	return configPath
}

func (c *configRunner) read() error {
	var (
		err error
	)
	// the returned error from vc.ReadInConfig() does not support standard
	// golang error identity via errors.Is, so we are forced to assume
	// that if we get an error, it is viper.ConfigFileNotFoundError
	//
	sequence := []tryReadConfigFn{
		func() error {
			// don't need to do anything here, as we use the config
			// as originally requested.
			//
			return nil
		},
		func() error {
			// try the home path
			//
			c.vc.AddConfigPath(c.home)

			return nil
		},
		func() error {
			// not found in home, therefore export default to
			// home path, which has already been added in previous
			// attempt.
			//
			err = c.export()

			return err
		},
	}

	iterator := collections.ForwardRunIt[tryReadConfigFn, error](sequence, nil)
	each := func(fn tryReadConfigFn) error {
		if e := fn(); e != nil {
			return e
		}

		return c.vc.ReadInConfig()
	}
	while := func(_ tryReadConfigFn, e error) bool {
		return e != nil
	}
	iterator.RunAll(each, while)

	return err
}

func (c *configRunner) export() error {
	path := c.DefaultPath()
	file := filepath.Join(path, common.Definitions.Pixa.ConfigType)
	content := []byte(defaultConfig)

	if !c.vfs.FileExists(file) {
		if err := c.vfs.MkdirAll(path, common.Permissions.Write); err != nil {
			return err
		}

		return c.vfs.WriteFile(file, content, common.Permissions.Write)
	}

	return nil
}

func (c *configRunner) handleLangSetting(config configuration.ViperConfig) {
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
