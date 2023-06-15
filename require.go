package jz

import (
	js "github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
)

type Registry struct {
	*require.Registry

	native   map[string]ModuleLoader
	compiled map[string]*js.Program

	srcLoader     SourceLoader
	globalFolders []string
}

type RequireModule struct {
	requireModule *require.RequireModule

	r           *Registry
	runtime     *Runtime
	modules     map[string]*js.Object
	nodeModules map[string]*js.Object
}

func (r *RequireModule) require(call js.FunctionCall) js.Value {
	ret, err := r.Require(call.Argument(0).String())
	if err != nil {
		if _, ok := err.(*js.Exception); !ok {
			panic(r.runtime.NewGoError(err))
		}
		panic(err)
	}
	return ret
}

type SourceLoader = require.SourceLoader

func NewRegistryWithLoader(srcLoader SourceLoader) *Registry {
	return &Registry{
		Registry:  require.NewRegistryWithLoader(srcLoader),
		srcLoader: srcLoader,
	}
}

// Enable adds the require() function to the specified runtime.
func (r *Registry) Enable(runtime *Runtime) *RequireModule {
	requireModule := &RequireModule{
		requireModule: r.Registry.Enable(runtime.Runtime),
		runtime:       runtime,
		modules:       map[string]*Object{},
		r:             r,
	}
	runtime.Set("require", requireModule.require)
	return requireModule
}

// Require can be used to import modules from Go source (similar to JS require() function).
func (r *RequireModule) Require(name string) (ret Value, err error) {
	module, err := r.resolve(name)
	if err != nil {
		return r.requireModule.Require(name)
	}
	ret = module.Get("exports")
	return
}
