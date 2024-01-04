package helpers

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"

	"github.com/samber/lo"
	"github.com/snivilised/extendio/collections"

	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/extendio/xfs/utils"
)

const offset = 2
const tabSize = 2

type DirectoryTreeBuilder struct {
	vfs       storage.VirtualFS
	root      string
	full      string
	stack     *collections.Stack[string]
	indexPath string
	write     bool
	depth     int
	padding   string
	silent    bool
}

func (r *DirectoryTreeBuilder) read() (*Directory, error) {
	data, err := os.ReadFile(r.indexPath) // always read from real fs

	if err != nil {
		return nil, err
	}

	var tree Tree
	err = xml.Unmarshal(data, &tree)

	if err != nil {
		return nil, err
	}

	return &tree.Root, nil
}

func (r *DirectoryTreeBuilder) status(path string) string {
	return lo.Ternary(utils.Exists(path), "✔️", "❌")
}

func (r *DirectoryTreeBuilder) pad() string {
	return string(bytes.Repeat([]byte{' '}, (r.depth+offset)*tabSize))
}

func (r *DirectoryTreeBuilder) refill() string {
	segments := r.stack.Content()
	return filepath.Join(segments...)
}

func (r *DirectoryTreeBuilder) inc(name string) {
	r.stack.Push(name)
	r.full = r.refill()

	r.depth++
	r.padding = r.pad()
}

func (r *DirectoryTreeBuilder) dec() {
	_, _ = r.stack.Pop()
	r.full = r.refill()

	r.depth--
	r.padding = r.pad()
}

func (r *DirectoryTreeBuilder) show(path, indicator, name string) {
	if !r.silent {
		status := r.status(path)
		fmt.Printf("%v(depth: '%v') (%v) %v: -> '%v' (🌟 %v)\n",
			r.padding, r.depth, status, indicator, name, r.full,
		)
	}
}

func (r *DirectoryTreeBuilder) walk() error {
	fmt.Printf("\n🤖 re-generating tree at '%v'\n", r.root)

	top, err := r.read()

	if err != nil {
		return err
	}

	r.full = r.root

	return r.dir(*top)
}

func (r *DirectoryTreeBuilder) dir(dir Directory) error { //nolint:gocritic // performance is not an issue
	r.inc(dir.Name)

	_, dn := utils.SplitParent(dir.Name)

	if r.write {
		err := r.vfs.MkdirAll(r.full, os.ModePerm)

		if err != nil {
			return err
		}
	}

	r.show(r.full, "📂", dn)

	for _, directory := range dir.Directories {
		err := r.dir(directory)
		if err != nil {
			return err
		}
	}

	for _, file := range dir.Files {
		fp := Path(r.full, file.Name)

		if r.write {
			err := r.vfs.WriteFile(fp, []byte(file.Text), os.ModePerm)
			if err != nil {
				return err
			}
		}

		r.show(fp, "  📜", file.Name)
	}

	r.dec()

	return nil
}

type Tree struct {
	XMLName xml.Name  `xml:"tree"`
	Root    Directory `xml:"directory"`
}

type Directory struct {
	XMLName     xml.Name    `xml:"directory"`
	Name        string      `xml:"name,attr"`
	Files       []File      `xml:"file"`
	Directories []Directory `xml:"directory"`
}

type File struct {
	XMLName xml.Name `xml:"file"`
	Name    string   `xml:"name,attr"`
	Text    string   `xml:",chardata"`
}

const doWrite = true

func Scientist(vfs storage.VirtualFS, index string, silent bool) string {
	repo := Repo("")
	research := filepath.Join(repo, "test", "data", "research")
	scientist := filepath.Join(research, "scientist")
	indexPath := filepath.Join(research, index)
	utils.Must(ensure(scientist, indexPath, vfs, silent))

	return scientist
}

func ensure(root, indexPath string, vfs storage.VirtualFS, silent bool) error {
	if vfs.DirectoryExists(root) {
		return nil
	}

	parent, _ := utils.SplitParent(root)
	builder := DirectoryTreeBuilder{
		vfs:       vfs,
		root:      root,
		stack:     collections.NewStackWith([]string{parent}),
		indexPath: indexPath,
		write:     doWrite,
		silent:    silent,
	}

	return builder.walk()
}
