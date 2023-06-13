package jz

import (
	"errors"
	"fmt"
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"github.com/jvatic/goja-babel"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type ImportMap map[string]string

type Runtime struct {
	*goja.Runtime
	UseBabel      bool
	UseTypeScript bool
	ImportMap
	registry *require.Registry
}

func GetContent(url string) ([]byte, error) {
	fmt.Println("GetContent", url)
	if strings.Contains(url, "http://") || strings.Contains(url, "https://") {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		return io.ReadAll(resp.Body)
	}

	return os.ReadFile(url)
}

func New() *Runtime {
	vm := &Runtime{
		Runtime:       goja.New(),
		UseBabel:      true,
		UseTypeScript: true,
		ImportMap:     ImportMap{},
	}

	vm.registry = require.NewRegistryWithLoader(func(name string) ([]byte, error) {
		fmt.Println(name)
		for importName, importURL := range vm.ImportMap {
			if name != "node_modules/"+importName {
				continue
			}
			b, err := GetContent(importURL)
			//fmt.Println("GetContent after", string(b))
			if err != nil {
				return nil, err
			}

			src := string(b)
			fmt.Println("goja.Compile", importURL)
			_, err = goja.Compile(importURL, src, false)
			if err != nil {
				log.Println(err)
				fmt.Println("Transform", importURL)
				return Transform(src)
			}
			return b, err
		}
		return nil, errors.New("Module does not exist")
	})
	vm.registry.Enable(vm.Runtime)
	console.Enable(vm.Runtime)

	return vm
}

func (r *Runtime) AddImportMap(m ImportMap) {
	for key, value := range m {
		r.ImportMap[key] = value
	}
}

func (r *Runtime) ClearImportMap() {
	r.ImportMap = ImportMap{}
}

func Transform(src string) ([]byte, error) {
	res, err := babel.Transform(strings.NewReader(src), map[string]interface{}{
		"presets": []string{"es2015"},
	})
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(res)

	//fmt.Println(string(b))

	return b, err
}

func (r *Runtime) RunScript(name, src string) (goja.Value, error) {
	b, err := Transform(src)
	if err != nil {
		return nil, err
	}
	src = string(b)

	//fmt.Println(src)

	return r.Runtime.RunScript(name, src)
}

func init() {
	babel.Init(4) // Setup 4 transformers (can be any number > 0)
}
