package jz

import (
	"errors"
	js "github.com/dop251/goja"
	"github.com/dop251/goja/parser"
	"github.com/dop251/goja_nodejs/require"
	"path"
	"text/template"
)

var (
	InvalidModuleError     = errors.New("Invalid module")
	IllegalModuleNameError = errors.New("Illegal module name")

	ModuleFileDoesNotExistError = errors.New("module file does not exist")
)

type ModuleLoader func(*js.Runtime, *js.Object)

var native map[string]ModuleLoader

func filepathClean(p string) string {
	return path.Clean(p)
}

func (r *Registry) getSource(p string) ([]byte, error) {
	srcLoader := r.srcLoader
	if srcLoader == nil {
		srcLoader = require.DefaultSourceLoader
	}
	return srcLoader(p)
}

func (r *Registry) getCompiledSource(p string) (*js.Program, error) {
	r.Lock()
	defer r.Unlock()

	prg := r.compiled[p]
	if prg == nil {
		buf, err := r.getSource(p)
		if err != nil {
			return nil, err
		}
		s := string(buf)

		if path.Ext(p) == ".json" {
			s = "module.exports = JSON.parse('" + template.JSEscapeString(s) + "')"
		}

		source := "(function(exports, require, module) {" + s + "\n})"
		parsed, err := js.Parse(p, source, parser.WithSourceMapLoader(r.srcLoader))
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

type Option func(*Registry)

func NewRegistry(opts ...Option) *Registry {
	r := &Registry{}

	for _, opt := range opts {
		opt(r)
	}

	return r
}
