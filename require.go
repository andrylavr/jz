package jz

import (
	"fmt"
	js "github.com/dop251/goja"
	"log"
	"path"
	"strings"
	"sync"
	"text/template"
)

type JSModule = *Object

type Require struct {
	modules map[string]Value
	vm      *Runtime
	sync.Mutex
	compiled map[string]*js.Program
}

func NewRequire(vm *Runtime) Value {
	require := &Require{
		modules: map[string]Value{},
		vm:      vm,
	}
	return vm.ToValue(require.require)
}

func notOk(err error) bool {
	if err != nil {
		log.Println(err)
		return true
	}
	return false
}

func (require *Require) require(name string) Value {
	fmt.Println("require", name)

	if m, ok := require.modules[name]; ok {
		fmt.Println("already", name)
		return m
	}

	resolvedURL, err := require.resolveURL(name)
	if notOk(err) {
		return require.vm.NewGoError(err)
	}

	fmt.Println("GetContent", name, resolvedURL)
	src, err := require.vm.getContent(resolvedURL)
	if notOk(err) {
		return require.vm.NewGoError(err)
	}

	fmt.Println("Transform", name, resolvedURL)
	src, err = require.vm.Transform(src)
	if notOk(err) {
		return require.vm.NewGoError(err)
	}

	fmt.Println("getCompiledSource", name, resolvedURL)
	prg, err := require.getCompiledSource(resolvedURL, src)
	if notOk(err) {
		return require.vm.NewGoError(err)
	}

	fmt.Println("RunProgram", name, resolvedURL)
	f, err := require.vm.RunProgram(prg)
	if notOk(err) {
		return require.vm.NewGoError(err)
	}

	fmt.Println("AssertFunction", name, resolvedURL)
	call, ok := js.AssertFunction(f)
	if !ok {
		err = fmt.Errorf("invalid module with name %s", name)
		notOk(err)
		return require.vm.NewGoError(err)
	}

	jsModule := require.createModuleObject(resolvedURL)
	jsExports := jsModule.Get("exports")
	jsRequire := require.vm.Get("require")

	// Run the module source, with "jsExports" as "this",
	// "jsExports" as the "exports" variable, "jsRequire"
	// as the "require" variable and "jsModule" as the
	// "module" variable (Nodejs capable).
	fmt.Println("call module function", name, resolvedURL)

	_, err = call(jsExports, jsExports, jsRequire, jsModule)
	if err != nil {
		return require.vm.NewGoError(err)
	}

	fmt.Println("ready", name, resolvedURL)
	fmt.Println()

	require.modules[name] = jsExports
	return jsExports
}

func (require *Require) resolveURL(name string) (string, error) {
	if importURL, ok := require.vm.ImportMap[name]; ok {
		return importURL, nil
	} else {
		for importName, importURL := range require.vm.ImportMap {
			if !strings.HasSuffix(importName, "/") {
				continue
			}
			if !strings.HasPrefix(name, importName) {
				continue
			}
			return strings.Replace(name, importName, importURL, 1), nil
		}
	}

	return "", fmt.Errorf("module with name %s is not found", name)
}

type Module struct {
	require   *Require
	Exports   *Object
	moduleURL string
}

func (m *Module) Require(name string) Value {
	fmt.Println("Module.Require", name, "moduleURL", m.moduleURL)
	return m.require.require(name)
}

func (r *Require) createModuleObject(moduleURL string) *Object {
	//module := r.vm.NewObject()
	//module.Set("exports", r.vm.NewObject())
	//module.Set("require", r.vm.NewObject())
	//return module

	module := &Module{
		require:   r,
		Exports:   r.vm.NewObject(),
		moduleURL: moduleURL,
	}

	return r.vm.ToValue(module).ToObject(r.vm.Runtime)
}

func (r *Require) getCompiledSource(p string, s string) (*js.Program, error) {
	r.Lock()
	defer r.Unlock()

	prg := r.compiled[p]
	if prg == nil {
		if path.Ext(p) == ".json" {
			s = "module.exports = JSON.parse('" + template.JSEscapeString(s) + "')"
		}

		source := "(function(exports, require, module) {" + s + "\n})"
		//parsed, err := js.Parse(p, source, parser.WithSourceMapLoader(r.srcLoader))
		parsed, err := js.Parse(p, source)
		if err != nil {
			return nil, err
		}
		prg, err = js.CompileAST(parsed, false)
		if err == nil {
			if r.compiled == nil {
				r.compiled = make(map[string]*js.Program)
			}
			r.compiled[p] = prg
		}
		return prg, err
	}
	return prg, nil
}
