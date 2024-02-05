package cfg_test

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"go.uber.org/mock/gomock"

	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/cobrass/src/assistant/mocks"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/pixa/src/app/cfg"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/snivilised/pixa/src/internal/helpers"
)

var (
	sourceID        = "github.com/snivilised/pixa"
	applicationName = "pixa"
	environment     = "PIXA_HOME"
)

type runnerTE struct {
	given   string
	should  string
	path    string
	arrange func(entry *runnerTE, path string)
	created func(entry *runnerTE, runner common.ConfigRunner)
	assert  func(entry *runnerTE, runner common.ConfigRunner, err error)
}

var _ = Describe("ConfigRunner", func() {

	var (
		repo       string
		configPath string
		vfs        storage.VirtualFS
		vc         *configuration.GlobalViperConfig
		ctrl       *gomock.Controller
		mock       *mocks.MockViperConfig
	)

	BeforeEach(func() {
		viper.Reset()
		vfs = storage.UseMemFS()
		vc = &configuration.GlobalViperConfig{}
		ctrl = gomock.NewController(GinkgoT())
		mock = mocks.NewMockViperConfig(ctrl)
		repo = helpers.Repo("")
		configPath = helpers.Path(repo, "test/data/configuration")
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	DescribeTable("",
		func(entry *runnerTE) {
			ci := common.ConfigInfo{
				Name:       "pixa",
				ConfigType: "yaml",
				ConfigPath: entry.path,
				Viper:      mock,
			}

			mock.EXPECT().SetConfigName(ci.Name).Do(func(n string) {
				vc.SetConfigName(n)
			}).AnyTimes()

			mock.EXPECT().SetConfigType(ci.ConfigType).Do(func(t string) {
				vc.SetConfigType(t)
			}).AnyTimes()

			mock.EXPECT().AutomaticEnv().AnyTimes()
			entry.arrange(entry, configPath)
			mock.EXPECT().InConfig(gomock.Any()).AnyTimes()
			mock.EXPECT().GetString(gomock.Any()).AnyTimes()

			runner, err := cfg.New(&ci, sourceID, applicationName, vfs)
			if entry.created != nil {
				entry.created(entry, runner)
			}

			if err == nil {
				err = runner.Run()
			}

			entry.assert(entry, runner, err)
		},
		func(entry *runnerTE) string {
			return fmt.Sprintf("🧪 ===> given: '%v', should: '%v'",
				entry.given, entry.should,
			)
		},

		Entry(nil, &runnerTE{
			given:  "config file present at PIXA_HOME",
			should: "use config at PIXA_HOME",
			arrange: func(entry *runnerTE, path string) {
				mock.EXPECT().ReadInConfig().Times(1)
				mock.EXPECT().AddConfigPath(path).Do(func(_ string) {
					vc.AddConfigPath(path)
				}).AnyTimes()

				mock.EXPECT().Get(gomock.Eq(environment)).DoAndReturn(func(_ string) string {
					return path
				}).AnyTimes()
			},
			assert: func(entry *runnerTE, runner common.ConfigRunner, err error) {
				Expect(err).Error().To(BeNil())
				Expect(runner).NotTo(BeNil())
			},
		}),

		Entry(nil, &runnerTE{
			given:  "config file present as configured by client, PIXA_HOME not defined",
			should: "use config at specified path",
			arrange: func(entry *runnerTE, path string) {
				mock.EXPECT().ReadInConfig().Times(1)
				mock.EXPECT().AddConfigPath(gomock.Any()).Do(func(_ string) {
					vc.AddConfigPath(path)
				}).AnyTimes()

				mock.EXPECT().Get(gomock.Eq(environment)).DoAndReturn(func(e string) string {
					return ""
				}).AnyTimes()
			},
			assert: func(entry *runnerTE, runner common.ConfigRunner, err error) {
				Expect(err).Error().To(BeNil())
				Expect(runner).NotTo(BeNil())
			},
		}),

		Entry(nil, &runnerTE{
			given:  "config file missing, but at default location, PIXA_HOME not defined",
			should: "use config at default location",
			arrange: func(entry *runnerTE, path string) {
				mock.EXPECT().Get(gomock.Eq(environment)).DoAndReturn(func(_ string) string {
					return ""
				}).AnyTimes()

				mock.EXPECT().ReadInConfig().Times(1).DoAndReturn(func() error {
					return viper.ConfigFileNotFoundError{}
				})

				mock.EXPECT().AddConfigPath(gomock.Any()).Do(func(_ string) {
					vc.AddConfigPath(path)
				}).AnyTimes()

				mock.EXPECT().ReadInConfig().Times(1).DoAndReturn(func() error {
					return nil
				})
			},
			created: func(_ *runnerTE, runner common.ConfigRunner) {
				name := fmt.Sprintf("%v.%v", helpers.PixaConfigTestFilename, helpers.PixaConfigType)
				path := filepath.Join(runner.DefaultPath(), name)
				content := []byte(cfg.GetDefaultConfigContent())

				_ = vfs.WriteFile(path, content, cfg.WritePerm)
			},
			assert: func(entry *runnerTE, runner common.ConfigRunner, err error) {
				Expect(err).Error().To(BeNil())
				Expect(runner).NotTo(BeNil())
			},
		}),

		Entry(nil, &runnerTE{
			given:  "config file completely missing",
			should: "use default exported config",
			arrange: func(entry *runnerTE, path string) {
				mock.EXPECT().Get(gomock.Eq(environment)).DoAndReturn(func(_ string) string {
					return ""
				}).AnyTimes()

				mock.EXPECT().ReadInConfig().Times(2).DoAndReturn(func() error {
					return viper.ConfigFileNotFoundError{}
				})

				mock.EXPECT().AddConfigPath(gomock.Any()).Do(func(_ string) {
					vc.AddConfigPath(path)
				}).AnyTimes()

				mock.EXPECT().ReadInConfig().Times(1).DoAndReturn(func() error {
					return nil
				})
			},
			assert: func(entry *runnerTE, runner common.ConfigRunner, err error) {
				Expect(err).Error().To(BeNil())
				Expect(runner).NotTo(BeNil())
			},
		}),
	)
})
