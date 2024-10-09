package lab

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing/fstest"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	ci18n "github.com/snivilised/cobrass/src/assistant/i18n"
	"github.com/snivilised/li18ngo"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/snivilised/pixa/src/internal/matchers"
	"github.com/snivilised/pixa/src/locale"
	"github.com/snivilised/traverse/collections"
	"github.com/snivilised/traverse/lfs"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
)

// Path creates a path from the parent combined with the relative path. The relative
// path is a file system path so should only contain forward slashes, not the standard
// file path separator as denoted by filepath.Separator, typically used when interacting
// with the local file system. Do not use trailing "/".
func Path(parent, relative string) string {
	if relative == "" {
		return parent
	}

	return parent + "/" + relative
}

// Repo gets the path of the repo with relative joined on
func Repo(relative string) string {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, _ := cmd.Output()
	repo := strings.TrimSpace(string(output))

	return Path(repo, relative)
}

func Normalise(p string) string {
	return strings.ReplaceAll(p, "/", string(filepath.Separator))
}

func Because(name, because string) string {
	return fmt.Sprintf("❌ for item named: '%v', because: '%v'", name, because)
}

func Reason(name string) string {
	return fmt.Sprintf("❌ for item named: '%v'", name)
}

func Log() string {
	return Repo("Test/test.log")
}

func UseI18n(l10nPath string) error {
	return li18ngo.Use(func(uo *li18ngo.UseOptions) {
		uo.From = li18ngo.LoadFrom{
			Path: l10nPath,
			Sources: li18ngo.TranslationFiles{
				locale.PixaSourceID: li18ngo.TranslationSource{
					Name: "dummy-cobrass",
				},

				ci18n.CobrassSourceID: li18ngo.TranslationSource{
					Name: "dummy-cobrass",
				},
			},
		}
	})
}

func ReadGlobalConfig(configPath string) error {
	var (
		err error
	)

	config := &configuration.GlobalViperConfig{}

	config.SetConfigType(common.Definitions.Pixa.ConfigType)
	config.SetConfigName(common.Definitions.Pixa.ConfigTestFilename)
	config.AddConfigPath(configPath)

	if e := config.ReadInConfig(); e != nil {
		err = errors.Wrap(e, "can't read config")
	}

	return err
}

// Must
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func SetupTest(
	index, configPath, l10nPath string,
	silent bool,
) (tfs lfs.TraverseFS, root string) {
	var (
		err error
	)

	viper.Reset()

	tfs, root = ResetFS(index, silent)

	if err = MockConfigFile(tfs, configPath); err != nil {
		ginkgo.Fail(err.Error())
	}

	if err = ReadGlobalConfig(configPath); err != nil {
		ginkgo.Fail(err.Error())
	}

	if err = UseI18n(l10nPath); err != nil {
		ginkgo.Fail(err.Error())
	}

	return tfs, root
}

// MockConfigFile create a dummy config file in the file system specified
func MockConfigFile(tfs lfs.TraverseFS, configPath string) error {
	var (
		err error
	)

	_ = tfs.MkDirAll(configPath, common.Permissions.Beezledub)

	if _, err = tfs.Create(
		filepath.Join(configPath, common.Definitions.Pixa.ConfigTestFilename),
	); err != nil {
		ginkgo.Fail(fmt.Sprintf("🔥 can't create dummy config (err: '%v')", err))
	}

	gomega.Expect(matchers.AsDirectory(configPath)).To(matchers.ExistInFS(tfs))

	return err
}

func ResetFS(index string, silent bool) (tfs lfs.TraverseFS, root string) {
	tfs = &TestTraverseFS{
		fstest.MapFS{
			".": &fstest.MapFile{
				Mode: os.ModeDir,
			},
		},
	}

	root = Scientist(tfs, index, silent)
	gomega.Expect(matchers.AsDirectory(root)).To(matchers.ExistInFS(tfs))

	return tfs, root
}

func Scientist(tfs lfs.TraverseFS, index string, silent bool) string {
	research := filepath.Join("test", "data", "research")
	scientist := filepath.Join(research, "scientist")
	indexPath := filepath.Join(research, index)
	Must(ensure(scientist, indexPath, tfs, silent))

	return scientist
}

// ensure
// func ensure(root, index string, provider *IOProvider, verbose bool) error {
// 	parent, _ := lfs.SplitParent(root)
// 	builder := directoryTreeBuilder{
// 		root:     TrimRoot(root),
// 		stack:    collections.NewStackWith([]string{parent}),
// 		index:    index,
// 		doWrite:  doWrite,
// 		provider: provider,
// 		verbose:  verbose,
// 		show: func(path string, exists existsEntry) {
// 			if !verbose {
// 				return
// 			}

// 			status := lo.Ternary(exists(path), "✅", "❌")

// 			fmt.Printf("---> %v path: '%v'\n", status, path)
// 		},
// 	}

// 	return builder.walk()
// }

func ensure(root, indexPath string, tfs lfs.TraverseFS, silent bool) error {
	_ = indexPath
	_ = silent

	if tfs.DirectoryExists(root) {
		return nil
	}

	parent, _ := lfs.SplitParent(root)
	builder := directoryTreeBuilder{
		// vfs:       tfs,
		root:  root,
		stack: collections.NewStackWith([]string{parent}),
		// indexPath: indexPath,
		// write:     true,
		// silent:    silent,
	}

	return builder.walk()
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
