package file

import (
	"errors"
	"path/filepath"
	"regexp"

	"github.com/jacokoo/fff/executor"
)

var (
	loaders   []Loader
	loaderMap = make(map[string]Loader)
)

func parsePath(pp string) Path {
	p := "@file://" + pp
	tokens := regexp.MustCompile(`@(\w+)://`).FindAllStringSubmatchIndex(p, -1)
	re := make([]*pathItem, 0)
	for i, v := range tokens {
		end := len(p)
		if i < len(tokens)-1 {
			end = tokens[i+1][0]
		}
		name, arg := p[v[2]:v[3]], p[v[1]:end]
		re = append(re, &pathItem{loaderMap[name], arg})
	}
	return &path{re}
}

func loadFile(file File) executor.Future {
	for _, v := range loaders {
		if v.Support(file) {
			return v.Create(file)
		}
	}
	return executor.ErrorFuture(errors.New("can not open file"))
}

func registerLoader(loader Loader) {
	loaderMap[loader.Name()] = loader
	loaders = append([]Loader{loader}, loaders...)
}

type localLoader struct {
}

func (fp *localLoader) Name() string                { return "file" }
func (fp *localLoader) Support(item File) bool      { return item.IsDir() }
func (fp *localLoader) Create(File) executor.Future { return executor.OkFuture(Root) }
func (fp *localLoader) PathDir(path string) string  { return filepath.Dir(path) }

func init() {
	registerLoader(new(localLoader))
}
