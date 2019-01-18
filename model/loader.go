package model

import (
	"errors"
	"os"
	"regexp"
)

// Loader such as ssh, zip, tar, tgz
type Loader interface {
	Name() string
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
}

type localLoader struct {
	root FileItem
}

func (fp *localLoader) Name() string                             { return "file" }
func (fp *localLoader) Support(item FileItem) bool               { return item.IsDir() }
func (fp *localLoader) Create(parent FileItem) (FileItem, error) { return fp.root, nil }

func registerLoader(loader Loader) {
	loaderMap[loader.Name()] = loader
	loaders = append([]Loader{loader}, loaders...)
}

// Load file item from path
// /a/b/c.zip@zip:///hello/path
// /a/b/c.fff@ssh:///opt/c.tgz@tgz:///hello/path
func Load(path string) (FileItem, error) {
	p := "@file://" + path
	tokens := regexp.MustCompile(`@(\w+)://`).FindAllStringSubmatchIndex(p, -1)
	var item FileItem
	for i, v := range tokens {
		end := len(p)
		if i < len(tokens)-1 {
			end = tokens[i+1][0]
		}

		name, arg := p[v[2]:v[3]], p[v[1]:end]
		i, err := loaderMap[name].Create(item)
		if err != nil {
			return nil, err
		}
		item, err = i.(DirOp).To(arg)
		if err != nil {
			return nil, err
		}
	}
	return item, nil
}

// LoadFile load a file as dir
func LoadFile(item FileItem) (FileItem, error) {
	for _, v := range loaders {
		if v.Support(item) {
			return v.Create(item)
		}
	}
	return nil, errors.New("can not open file")
}
