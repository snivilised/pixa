package helpers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/spf13/viper"

	"github.com/pkg/errors"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	ci18n "github.com/snivilised/cobrass/src/assistant/i18n"
	cmocks "github.com/snivilised/cobrass/src/assistant/mocks"
	"github.com/snivilised/pixa/src/app/mocks"
	"github.com/snivilised/pixa/src/cfg"
	"github.com/snivilised/pixa/src/i18n"
	"github.com/snivilised/pixa/src/internal/matchers"

	xi18n "github.com/snivilised/extendio/i18n"
	"github.com/snivilised/extendio/xfs/storage"
	"golang.org/x/text/language"
)

const (
	PixaConfigTestFilename = "pixa-test"
	PixaConfigType         = "yml"
	ShrinkCommandName      = "shrink"
	ProgName               = "magick"
	Faydeaudeau            = os.FileMode(0o777)
	Beezledub              = os.FileMode(0o666)
	Silent                 = true
	Verbose                = false
)

func Path(parent, relative string) string {
	segments := strings.Split(relative, "/")
	return filepath.Join(append([]string{parent}, segments...)...)
}

func Normalise(p string) string {
	return strings.ReplaceAll(p, "/", string(filepath.Separator))
}

func Reason(name string) string {
	return fmt.Sprintf("❌ for item named: '%v'", name)
}

func JoinCwd(segments ...string) string {
	if current, err := os.Getwd(); err == nil {
		parent, _ := filepath.Split(current)
		grand := filepath.Dir(parent)
		great := filepath.Dir(grand)
		all := append([]string{great}, segments...)

		return filepath.Join(all...)
	}

	panic("could not get root path")
}

func Root() string {
	if current, err := os.Getwd(); err == nil {
		return current
	}

	panic("could not get root path")
}

func Repo(relative string) string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	if bytes, err := cmd.Output(); err != nil {
		panic(errors.Wrap(err, "couldn't get repo root"))
	} else {
		segments := strings.Split(relative, "/")
		output := strings.TrimSuffix(string(bytes), "\n")
		path := []string{output}
		path = append(path, segments...)

		return filepath.Join(path...)
	}
}

func Log() string {
	if current, err := os.Getwd(); err == nil {
		parent, _ := filepath.Split(current)
		grand := filepath.Dir(parent)
		great := filepath.Dir(grand)

		return filepath.Join(great, "Test", "test.log")
	}

	panic("could not get root path")
}

func SetupTest(
	index, configPath, l10nPath string,
	silent bool,
) (vfs storage.VirtualFS, root string, config configuration.ViperConfig) {
	var (
		err error
	)

	viper.Reset()

	vfs, root = ResetFS(index, silent)

	if err = MockConfigFile(vfs, configPath); err != nil {
		ginkgo.Fail(err.Error())
	}

	if config, err = ReadGlobalConfig(configPath); err != nil {
		ginkgo.Fail(err.Error())
	}

	if err = UseI18n(l10nPath); err != nil {
		ginkgo.Fail(err.Error())
	}

	return vfs, root, config
}

func UseI18n(l10nPath string) error {
	return xi18n.Use(func(uo *xi18n.UseOptions) {
		uo.From = xi18n.LoadFrom{
			Path: l10nPath,
			Sources: xi18n.TranslationFiles{
				i18n.PixaSourceID: xi18n.TranslationSource{
					Name: "dummy-cobrass",
				},

				ci18n.CobrassSourceID: xi18n.TranslationSource{
					Name: "dummy-cobrass",
				},
			},
		}
	})
}

func ReadGlobalConfig(configPath string) (*configuration.GlobalViperConfig, error) {
	var (
		err error
	)

	config := &configuration.GlobalViperConfig{}

	config.SetConfigType(PixaConfigType)
	config.SetConfigName(PixaConfigTestFilename)
	config.AddConfigPath(configPath)

	if e := config.ReadInConfig(); e != nil {
		err = errors.Wrap(e, "can't read config")
	}

	return config, err
}

// MockConfigFile create a dummy config file in the file system specified
func MockConfigFile(vfs storage.VirtualFS, configPath string) error {
	var (
		err error
	)

	_ = vfs.MkdirAll(configPath, Beezledub)

	if _, err = vfs.Create(filepath.Join(configPath, PixaConfigTestFilename)); err != nil {
		ginkgo.Fail(fmt.Sprintf("🔥 can't create dummy config (err: '%v')", err))
	}

	gomega.Expect(matchers.AsDirectory(configPath)).To(matchers.ExistInFS(vfs))

	return err
}

func DoMockConfigs(
	config configuration.ViperConfig,
	profilesReader *mocks.MockProfilesConfigReader,
	schemesReader *mocks.MockSchemesConfigReader,
	samplerReader *mocks.MockSamplerConfigReader,
	mockAdvancedReader *mocks.MockAdvancedConfigReader,
	mockLoggingReader *mocks.MockLoggingConfigReader,
) {
	DoMockProfilesConfigsWith(ProfilesConfigData, config, profilesReader)
	DoMockSchemesConfigWith(SchemesConfigData, config, schemesReader)
	DoMockSamplerConfigWith(SamplerConfigData, config, samplerReader)
	DoMockAdvancedConfigWith(AdvancedConfigData, config, mockAdvancedReader)
	DoMockLoggingConfigWith(LoggingConfigData, config, mockLoggingReader)
}

func DoMockReadInConfig(config *cmocks.MockViperConfig) {
	config.EXPECT().ReadInConfig().DoAndReturn(
		func() error {
			return nil
		},
	).AnyTimes()
}

func DoMockProfilesConfigsWith(
	data cfg.ProfilesConfigMap,
	config configuration.ViperConfig,
	reader *mocks.MockProfilesConfigReader,
) {
	reader.EXPECT().Read(config).DoAndReturn(
		func(viper configuration.ViperConfig) (cfg.ProfilesConfig, error) {
			stub := &cfg.MsProfilesConfig{
				Profiles: data,
			}

			return stub, nil
		},
	).AnyTimes()
}

func DoMockSchemesConfigWith(
	data cfg.SchemesConfig,
	config configuration.ViperConfig,
	reader *mocks.MockSchemesConfigReader,
) {
	reader.EXPECT().Read(config).DoAndReturn(
		func(viper configuration.ViperConfig) (cfg.SchemesConfig, error) {
			stub := data

			return stub, nil
		},
	).AnyTimes()
}

func DoMockSamplerConfigWith(
	data cfg.SamplerConfig,
	config configuration.ViperConfig,
	reader *mocks.MockSamplerConfigReader,
) {
	reader.EXPECT().Read(config).DoAndReturn(
		func(viper configuration.ViperConfig) (cfg.SamplerConfig, error) {
			stub := data

			return stub, nil
		},
	).AnyTimes()
}

func DoMockAdvancedConfigWith(
	data cfg.AdvancedConfig,
	config configuration.ViperConfig,
	reader *mocks.MockAdvancedConfigReader,
) {
	reader.EXPECT().Read(config).DoAndReturn(
		func(viper configuration.ViperConfig) (cfg.AdvancedConfig, error) {
			stub := data

			return stub, nil
		},
	).AnyTimes()
}

func DoMockLoggingConfigWith(
	data cfg.LoggingConfig,
	config configuration.ViperConfig,
	reader *mocks.MockLoggingConfigReader,
) {
	reader.EXPECT().Read(config).DoAndReturn(
		func(viper configuration.ViperConfig) (cfg.LoggingConfig, error) {
			stub := data

			return stub, nil
		},
	).AnyTimes()
}

func ResetFS(index string, silent bool) (vfs storage.VirtualFS, root string) {
	vfs = storage.UseMemFS()
	root = Scientist(vfs, index, silent)
	gomega.Expect(matchers.AsDirectory(root)).To(matchers.ExistInFS(vfs))

	return vfs, root
}

type DetectorStub struct {
}

func (j *DetectorStub) Scan() language.Tag {
	return language.BritishEnglish
}

type ExecutorStub struct {
	Name string
}

func (e *ExecutorStub) ProgName() string {
	return e.Name
}

func (e *ExecutorStub) Look() (string, error) {
	return "", nil
}

func (e *ExecutorStub) Execute(_ ...string) error {
	return nil
}

type DirectoryQuantities struct {
	Files   uint
	Folders uint
}
