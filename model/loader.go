package model

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// Loader such as ssh, zip, tar, tgz
type Loader interface {
	Name() string
	Seperator() string
	Support(FileItem) bool
	Create(FileItem) (FileItem, error)
}

var (
	loaders   []Loader
	loaderMap = make(map[string]Loader)
)

func init() {
	fi, _ := os.Stat("/")
	v := newFile("", fi)
	root := &dir{v, &defaultDirOp{&defaultOp{v}}}
	local := &localLoader{root}
	registerLoader(local)
	registerLoader(new(zipLoader))
	registerLoader(new(tarLoader))
	registerLoader(new(tgzLoader))
	registerLoader(new(sshLoader))
}

// LoaderString return @xxx://
func LoaderString(loader Loader) string {
	return fmt.Sprintf("@%s://", loader.Name())
}

type localLoader struct {
	root FileItem
}

func (fp *localLoader) Name() string                             { return "file" }
func (fp *localLoader) Seperator() string                        { return string(filepath.Separator) }
func (fp *localLoader) Support(item FileItem) bool               { return item.IsDir() }
func (fp *localLoader) Create(parent FileItem) (FileItem, error) { return fp.root, nil }

func registerLoader(loader Loader) {
	loaderMap[loader.Name()] = loader
	loaders = append([]Loader{loader}, loaders...)
}

// PathItem parse result
type PathItem struct {
	Loader, Path, Seperator string
}

// ParsePath parse path
// /a/b/c.zip@zip:///hello/path
// /a/b/c.fff@ssh:///opt/c.tgz@tgz:///hello/path
func ParsePath(path string) []*PathItem {
	p := "@file://" + path
	tokens := regexp.MustCompile(`@(\w+)://`).FindAllStringSubmatchIndex(p, -1)
	re := make([]*PathItem, 0)
	for i, v := range tokens {
		end := len(p)
		if i < len(tokens)-1 {
			end = tokens[i+1][0]
		}
		name, arg := p[v[2]:v[3]], p[v[1]:end]
		re = append(re, &PathItem{name, arg, loaderMap[name].Seperator()})
	}
	return re
}

// Load file item from path
func Load(path string) (FileItem, error) {
	var item FileItem
	pis := ParsePath(path)
	for _, p := range pis {
		i, err := loaderMap[p.Loader].Create(item)
		if err != nil {
			return nil, err
		}

		item, err = i.(DirOp).To(p.Path)
		if err != nil {
			return nil, err
		}
	}
	return item, nil
}

// LoadFile load a file as dir, such as .tar file
func LoadFile(item FileItem) (FileItem, error) {
	for _, v := range loaders {
		if v.Support(item) {
			return v.Create(item)
		}
	}
	return nil, errors.New("can not open file")
}
