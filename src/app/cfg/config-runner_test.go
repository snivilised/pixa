package cfg_test

import (
	"errors"
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // foo
	. "github.com/onsi/gomega"    //nolint:revive // foo
	"github.com/spf13/viper"
	"go.uber.org/mock/gomock"

	"github.com/snivilised/cobrass/src/assistant/mocks"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/li18ngo"
	"github.com/snivilised/pixa/src/app/cfg"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/snivilised/pixa/src/internal/helpers"
)

var (
	sourceID    = "github.com/snivilised/pixa"
	environment = "PIXA_HOME"
	useXDG      = ""
)

type testScope struct {
}

func (f *testScope) ConfigDirs() ([]string, error) {
	return []string{
		filepath.Join(string(filepath.Separator), "foo"),
		filepath.Join(string(filepath.Separator), "bar"),
	}, nil
}

func (f *testScope) LogPath(filename string) (string, error) {
	return filename, nil
}

type errorScope struct {
}

func (f *errorScope) ConfigDirs() ([]string, error) {
	return []string{}, errors.New("fake could not get config dirs")
}

func (f *errorScope) LogPath(filename string) (string, error) {
	return filename, nil
}

type runnerTE struct {
	given   string
	should  string
	path    string
	scope   common.ConfigScope
	arrange func(entry *runnerTE, path string)
	created func(entry *runnerTE, runner common.ConfigRunner)
	assert  func(entry *runnerTE, runner common.ConfigRunner, err error)
}

var _ = Describe("ConfigRunner", Ordered, func() {
	var (
		repo       string
		configPath string
		vfs        storage.VirtualFS
		ctrl       *gomock.Controller
		mock       *mocks.MockViperConfig
	)

	BeforeAll(func() {
		Expect(li18ngo.Use()).To(Succeed())
	})

	BeforeEach(func() {
		viper.Reset()
		vfs = storage.UseMemFS()
		ctrl = gomock.NewController(GinkgoT())
		mock = mocks.NewMockViperConfig(ctrl)
		repo = helpers.Repo("")
		configPath = helpers.Path(repo, "test/data/configuration")
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	DescribeTable("mocked",
		func(entry *runnerTE) {
			ci := common.ConfigInfo{
				Name:       common.Definitions.Pixa.AppName,
				ConfigType: common.Definitions.Pixa.ConfigType,
				ConfigPath: entry.path,
				Viper:      mock,
				Scope:      entry.scope,
			}
			// this is why I hate mocking, requires too much
			// knowledge of the implementation, making the tests
			// brittle, but its necessary for this use case
			//
			mock.EXPECT().SetConfigName(ci.Name).AnyTimes()
			mock.EXPECT().SetConfigType(ci.ConfigType).AnyTimes()
			mock.EXPECT().AutomaticEnv().AnyTimes()

			entry.arrange(entry, configPath)

			mock.EXPECT().InConfig(gomock.Any()).AnyTimes()
			mock.EXPECT().GetString(gomock.Any()).AnyTimes()

			runner, err := cfg.New(&ci, sourceID, common.Definitions.Pixa.AppName, vfs)
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
			arrange: func(_ *runnerTE, path string) {
				mock.EXPECT().Get(gomock.Eq(common.Definitions.Environment.UseXDG)).DoAndReturn(func(_ string) string {
					return ""
				}).AnyTimes()

				mock.EXPECT().ReadInConfig().Times(1)
				mock.EXPECT().AddConfigPath(path).AnyTimes()
				mock.EXPECT().Get(gomock.Eq(environment)).DoAndReturn(func(_ string) string {
					return path
				}).AnyTimes()
			},
			assert: func(_ *runnerTE, runner common.ConfigRunner, err error) {
				Expect(err).Error().To(BeNil())
				Expect(runner).NotTo(BeNil())
			},
		}),

		Entry(nil, &runnerTE{
			given:  "config file present as configured by client, PIXA_HOME not defined",
			should: "use config at specified path",
			arrange: func(_ *runnerTE, _ string) {
				mock.EXPECT().Get(gomock.Eq(common.Definitions.Environment.UseXDG)).DoAndReturn(func(_ string) string {
					return ""
				}).AnyTimes()

				mock.EXPECT().ReadInConfig().Times(1)
				mock.EXPECT().AddConfigPath(gomock.Any()).AnyTimes()
				mock.EXPECT().Get(gomock.Eq(environment)).DoAndReturn(func(_ string) string {
					return ""
				}).AnyTimes()
			},
			assert: func(_ *runnerTE, runner common.ConfigRunner, err error) {
				Expect(err).Error().To(BeNil())
				Expect(runner).NotTo(BeNil())
			},
		}),

		Entry(nil, &runnerTE{
			given:  "config file missing, but at default location, PIXA_HOME not defined",
			should: "use config at default location",
			arrange: func(_ *runnerTE, _ string) {
				mock.EXPECT().Get(gomock.Eq(common.Definitions.Environment.UseXDG)).DoAndReturn(func(_ string) string {
					return ""
				}).AnyTimes()

				mock.EXPECT().Get(gomock.Eq(environment)).DoAndReturn(func(_ string) string {
					return ""
				}).AnyTimes()

				mock.EXPECT().ReadInConfig().Times(1).DoAndReturn(func() error {
					return viper.ConfigFileNotFoundError{}
				})
				mock.EXPECT().AddConfigPath(gomock.Any()).AnyTimes()
				mock.EXPECT().ReadInConfig().Times(1).DoAndReturn(func() error {
					return nil
				})
			},
			created: func(_ *runnerTE, runner common.ConfigRunner) {
				name := fmt.Sprintf("%v.%v",
					common.Definitions.Pixa.ConfigTestFilename,
					common.Definitions.Pixa.ConfigType,
				)
				path := filepath.Join(runner.DefaultPath(), name)
				content := []byte(cfg.GetDefaultConfigContent())

				_ = vfs.WriteFile(path, content, common.Permissions.Write)
			},
			assert: func(_ *runnerTE, runner common.ConfigRunner, err error) {
				Expect(err).Error().To(BeNil())
				Expect(runner).NotTo(BeNil())
			},
		}),

		Entry(nil, &runnerTE{
			given:  "config file completely missing",
			should: "use default exported config",
			arrange: func(_ *runnerTE, _ string) {
				mock.EXPECT().Get(gomock.Eq(common.Definitions.Environment.UseXDG)).DoAndReturn(func(_ string) string {
					return ""
				}).AnyTimes()

				mock.EXPECT().Get(gomock.Eq(environment)).DoAndReturn(func(_ string) string {
					return ""
				}).AnyTimes()

				mock.EXPECT().ReadInConfig().Times(2).DoAndReturn(func() error {
					return viper.ConfigFileNotFoundError{}
				})
				mock.EXPECT().AddConfigPath(gomock.Any()).AnyTimes()
				mock.EXPECT().ReadInConfig().Times(1).DoAndReturn(func() error {
					return nil
				})
			},
			assert: func(_ *runnerTE, runner common.ConfigRunner, err error) {
				Expect(err).Error().To(BeNil())
				Expect(runner).NotTo(BeNil())
			},
		}),

		Entry(nil, &runnerTE{
			given:  "use XDG, config file completely missing",
			should: "use default exported config",
			scope:  &testScope{},
			arrange: func(_ *runnerTE, _ string) {
				mock.EXPECT().Get(gomock.Eq(common.Definitions.Environment.UseXDG)).DoAndReturn(func(_ string) string {
					return "true"
				}).AnyTimes()

				mock.EXPECT().Get(gomock.Eq(environment)).DoAndReturn(func(_ string) string {
					return ""
				}).AnyTimes()

				mock.EXPECT().ReadInConfig().Times(2).DoAndReturn(func() error {
					return viper.ConfigFileNotFoundError{}
				})
				mock.EXPECT().AddConfigPath(gomock.Any()).AnyTimes()
				mock.EXPECT().ReadInConfig().Times(1).DoAndReturn(func() error {
					return nil
				})
			},
			assert: func(_ *runnerTE, runner common.ConfigRunner, err error) {
				Expect(err).Error().To(BeNil())
				Expect(runner).NotTo(BeNil())
			},
		}),

		Entry(nil, &runnerTE{
			given:  "scope returns error, config file completely missing",
			should: "use default exported config",
			scope:  &errorScope{},
			arrange: func(_ *runnerTE, _ string) {
				mock.EXPECT().Get(gomock.Eq(common.Definitions.Environment.UseXDG)).DoAndReturn(func(_ string) string {
					return ""
				}).AnyTimes()

				mock.EXPECT().Get(gomock.Eq(environment)).DoAndReturn(func(_ string) string {
					return ""
				}).AnyTimes()

				mock.EXPECT().ReadInConfig().Times(2).DoAndReturn(func() error {
					return viper.ConfigFileNotFoundError{}
				})
				mock.EXPECT().AddConfigPath(gomock.Any()).AnyTimes()
				mock.EXPECT().ReadInConfig().Times(1).DoAndReturn(func() error {
					return nil
				})
			},
			assert: func(_ *runnerTE, runner common.ConfigRunner, err error) {
				Expect(err).Error().To(BeNil())
				Expect(runner).NotTo(BeNil())
			},
		}),
	)
})
