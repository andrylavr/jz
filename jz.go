package jz

import (
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/jvatic/goja-babel"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type ImportMap map[string]string
type ModuleCache map[string][]byte
type Value = goja.Value
type Object = goja.Object

type Runtime struct {
	*goja.Runtime
	UseBabel      bool
	UseTypeScript bool
	ImportMap
	registry    *Registry
	moduleCache ModuleCache
}

func GetContent(url string) ([]byte, error) {
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
		moduleCache:   ModuleCache{},
	}

	vm.registry = NewRegistryWithLoader(func(path string) ([]byte, error) {
		b, err := GetContent(path)
		if err != nil {
			return nil, err
		}
		return vm.Transform(string(b))
	})

	vm.registry.Enable(vm)
	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())

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

// Transform transforms src from ES6|TS to ES5
func (r *Runtime) Transform(src string) ([]byte, error) {
	if !r.UseBabel {
		return []byte(src), nil
	}
	_, err := goja.Compile("", src, false)
	if err == nil {
		return []byte(src), err
	}

	options := map[string]interface{}{
		"presets": []string{"es2015"},
	}

	if r.UseTypeScript {
		options["plugins"] = []string{
			"transform-typescript",
		}
	}

	res, err := babel.Transform(strings.NewReader(src), options)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(res)

	//fmt.Println(string(b))

	return b, err
}

func (r *Runtime) RunScript(name, src string) (Value, error) {
	b, err := r.Transform(src)
	if err != nil {
		return nil, err
	}
	src = string(b)

	return r.Runtime.RunScript(name, src)
}

func (r *Runtime) RunFile(filename string) (Value, error) {
	b, err := GetContent(filename)
	if err != nil {
		return nil, err
	}
	src := string(b)
	b, err = r.Transform(src)
	if err != nil {
		return nil, err
	}
	src = string(b)

	return r.Runtime.RunScript(filename, src)
}

// RunString executes the given string in the global context.
func (r *Runtime) RunString(str string) (Value, error) {
	return r.RunScript("", str)
}

func init() {
	err := babel.Init(4) // Setup 4 transformers (can be any number > 0)
	if err != nil {
		log.Println(err)
	}
}
