package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	ci18n "github.com/snivilised/cobrass/src/assistant/i18n"
	"github.com/snivilised/pixa/src/i18n"

	xi18n "github.com/snivilised/extendio/i18n"
	"golang.org/x/text/language"
)

const (
	PixaConfigTestFilename = "pixa-test"
	PixaConfigType         = "yml"
	ShrinkCommandName      = "shrink"
	ProgName               = "magick"
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
	_, filename, _, _ := runtime.Caller(0) //nolint:dogsled // use of 3 _ is out of our control
	return Path(filepath.Dir(filename), relative)
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
