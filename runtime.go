package jz

import (
	"github.com/dop251/goja"
	"github.com/jvatic/goja-babel"
	"io"
	"log"
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
	//require         *Require
	TransformCacher Cacher
	ContentCacher   Cacher
}

func New() *Runtime {
	vm := &Runtime{
		Runtime:       goja.New(),
		UseBabel:      true,
		UseTypeScript: true,
		ImportMap:     ImportMap{},
	}

	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())

	vm.Set("console", NewConsole(vm))
	vm.Set("require", NewRequire(vm))

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
func (r *Runtime) Transform(src string) (string, error) {
	if r.TransformCacher != nil && r.TransformCacher.Exists(src) {
		return r.TransformCacher.Get(src), nil
	}

	if !r.UseBabel {
		return src, nil
	}
	_, err := goja.Compile("", src, false)
	if err == nil {
		return src, nil
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
		return "", err
	}
	b, err := io.ReadAll(res)
	srcTransformed := string(b)

	if err == nil && r.TransformCacher != nil {
		r.TransformCacher.Save(src, srcTransformed)
	}

	return srcTransformed, err
}

func (r *Runtime) RunScript(name, src string) (Value, error) {
	src, err := r.Transform(src)
	if err != nil {
		return nil, err
	}

	return r.Runtime.RunScript(name, src)
}

func (r *Runtime) RunFile(filename string) (Value, error) {
	src, err := r.getContent(filename)
	if err != nil {
		return nil, err
	}

	return r.RunScript(filename, src)
}

// RunString executes the given string in the global context.
func (r *Runtime) RunString(str string) (Value, error) {
	return r.RunScript("", str)
}

func (r *Runtime) getContent(url string) (string, error) {
	isHTTP := strings.Contains(url, "http://") || strings.Contains(url, "https://")
	if !isHTTP || r.ContentCacher == nil {
		return GetContent(url)
	}

	if r.ContentCacher.Exists(url) {
		return r.ContentCacher.Get(url), nil
	}
	content, err := GetContent(url)
	if err != nil {
		return "", err
	}
	r.ContentCacher.Save(url, content)
	return content, err
}

func init() {
	err := babel.Init(4) // Setup 4 transformers (can be any number > 0)
	if err != nil {
		log.Println(err)
	}
}
