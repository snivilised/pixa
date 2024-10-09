package lab

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing/fstest"

	"github.com/samber/lo"
	"github.com/snivilised/traverse/collections"
)

const (
	offset  = 2
	tabSize = 2
	doWrite = true
)

// TODO: should this be called science?
func Musico(verbose bool, portions ...string) (tsys *TestTraverseFS, root string) {
	tsys = &TestTraverseFS{
		fstest.MapFS{
			".": &fstest.MapFile{
				Mode: os.ModeDir,
			},
		},
	}

	return tsys, Provision(
		NewMemWriteProvider(tsys, os.ReadFile, portions...),
		verbose,
		portions...,
	)
}

func Provision(provider *IOProvider, verbose bool, portions ...string) (root string) {
	repo := Repo(filepath.Join("test", "data", "MUSICO"))

	if ensure(repo,
		"test/data/research/citizen-scientist-index.xml",
		provider.tfs, verbose,
	) != nil {
		return ""
	}

	if verbose {
		fmt.Printf("\n🤖 re-generated tree at '%v' (filters: '%v')\n\n",
			repo, strings.Join(portions, ", "),
		)
	}

	return repo
}

func TrimRoot(root string) string {
	// omit leading '/', because test-fs stupidly doesn't like it,
	// so we have to jump through hoops
	if strings.HasPrefix(root, string(filepath.Separator)) {
		return root[1:]
	}

	pattern := `^[a-zA-Z]:[\\/]*`
	re := regexp.MustCompile(pattern)

	return re.ReplaceAllString(root, "")
}

func NewMemWriteProvider(tfs *TestTraverseFS,
	indexReader readFile,
	portions ...string,
) *IOProvider {
	filter := lo.Ternary(len(portions) > 0,
		matcher(func(path string) bool {
			for _, portion := range portions {
				if strings.Contains(path, portion) {
					return true
				}
			}

			return false
		}),
		matcher(func(string) bool {
			return true
		}),
	)

	// PS: to check the existence of a path in an fs in production
	// code, use fs.Stat(fsys, path) instead of os.Stat/os.Lstat

	return &IOProvider{
		tfs:    tfs,
		filter: filter,
		file: fileHandler{
			in: indexReader,
			out: writeFile(func(name string, data []byte, mode os.FileMode, show display) error {
				if name == "" {
					return nil
				}

				if filter(name) {
					trimmed := TrimRoot(name)
					tfs.MapFS[trimmed] = &fstest.MapFile{
						Data: data,
						Mode: mode,
					}
					show(trimmed, func(path string) bool {
						entry, ok := tfs.MapFS[path]
						return ok && !entry.Mode.IsDir()
					})
				}

				return nil
			}),
		},
		folder: folderHandler{
			out: writeFolder(func(path string, mode os.FileMode, show display, isRoot bool) error {
				if path == "" {
					return nil
				}

				if isRoot || filter(path) {
					trimmed := TrimRoot(path)
					tfs.MapFS[trimmed] = &fstest.MapFile{
						Mode: mode | os.ModeDir,
					}
					show(trimmed, func(path string) bool {
						entry, ok := tfs.MapFS[path]
						return ok && entry.Mode.IsDir()
					})
				}

				return nil
			}),
		},
	}
}

type (
	entryExists interface {
		exists(path string) bool
	}

	existsEntry func(path string) bool

	display func(path string, exists existsEntry)

	fileReader interface {
		read(name string) ([]byte, error)
	}

	readFile func(name string) ([]byte, error)

	fileWriter interface {
		write(name string, data []byte, perm os.FileMode, show display) error
	}

	writeFile func(name string, data []byte, perm os.FileMode, show display) error

	folderWriter interface {
		write(path string, perm os.FileMode, show display, isRoot bool) error
	}

	writeFolder func(path string, perm os.FileMode, show display, isRoot bool) error

	filter interface {
		match(portion string) bool
	}

	matcher func(portion string) bool

	fileHandler struct {
		in  fileReader
		out fileWriter
	}

	folderHandler struct {
		out folderWriter
	}

	IOProvider struct {
		tfs    *TestTraverseFS
		filter filter
		file   fileHandler
		folder folderHandler
	}

	Tree struct {
		XMLName xml.Name  `xml:"tree"`
		Root    Directory `xml:"directory"`
	}

	Directory struct {
		XMLName     xml.Name    `xml:"directory"`
		Name        string      `xml:"name,attr"`
		Files       []File      `xml:"file"`
		Directories []Directory `xml:"directory"`
	}

	File struct {
		XMLName xml.Name `xml:"file"`
		Name    string   `xml:"name,attr"`
		Text    string   `xml:",chardata"`
	}
)

func (fn readFile) read(name string) ([]byte, error) {
	return fn(name)
}

func (fn writeFile) write(name string, data []byte, perm os.FileMode, show display) error {
	return fn(name, data, perm, show)
}

func (fn existsEntry) exists(path string) bool {
	return fn(path)
}

func (fn writeFolder) write(path string, perm os.FileMode, show display, isRoot bool) error {
	return fn(path, perm, show, isRoot)
}

func (fn matcher) match(portion string) bool {
	return fn(portion)
}

// directoryTreeBuilder
type directoryTreeBuilder struct {
	root     string
	full     string
	stack    *collections.Stack[string]
	index    string
	doWrite  bool
	depth    int
	padding  string
	provider *IOProvider
	verbose  bool
	show     display
}

func (r *directoryTreeBuilder) read() (*Directory, error) {
	data, err := r.provider.file.in.read(r.index)

	if err != nil {
		return nil, err
	}

	var tree Tree

	if ue := xml.Unmarshal(data, &tree); ue != nil {
		return nil, ue
	}

	return &tree.Root, nil
}

func (r *directoryTreeBuilder) pad() string {
	return string(bytes.Repeat([]byte{' '}, (r.depth+offset)*tabSize))
}

func (r *directoryTreeBuilder) refill() string {
	segments := r.stack.Content()
	return filepath.Join(segments...)
}

func (r *directoryTreeBuilder) inc(name string) {
	r.stack.Push(name)
	r.full = r.refill()

	r.depth++
	r.padding = r.pad()
}

func (r *directoryTreeBuilder) dec() {
	_, _ = r.stack.Pop()
	r.full = r.refill()

	r.depth--
	r.padding = r.pad()
}

func (r *directoryTreeBuilder) walk() error {
	top, err := r.read()

	if err != nil {
		return err
	}

	r.full = r.root

	return r.dir(*top, true)
}

func (r *directoryTreeBuilder) dir(dir Directory, isRoot bool) error { //nolint:gocritic // performance is not a concern
	r.inc(dir.Name)

	if r.doWrite {
		if err := r.provider.folder.out.write(
			r.full,
			os.ModePerm,
			r.show,
			isRoot,
		); err != nil {
			return err
		}
	}

	for _, directory := range dir.Directories {
		if err := r.dir(directory, false); err != nil {
			return err
		}
	}

	for _, file := range dir.Files {
		full := Path(r.full, file.Name)

		if r.doWrite {
			if err := r.provider.file.out.write(
				full,
				[]byte(file.Text),
				os.ModePerm,
				r.show,
			); err != nil {
				return err
			}
		}
	}

	r.dec()

	return nil
}
