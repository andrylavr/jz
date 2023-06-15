package jz_test

import (
	_ "embed"
	"github.com/andrylavr/jz"
	"testing"
)

//go:embed cmd/example1/example1.js
var example1 string

func TestJS(t *testing.T) {
	vm := jz.New()
	vm.ImportMap["three"] = "https://threejs.org/build/three.js"
	vm.ImportMap["three/addons/"] = "https://threejs.org/examples/jsm/"

	_, err := vm.RunScript("example1.js", example1)
	if err != nil {
		t.Error(err)
	}
}

//go:embed cmd/example1/example2.ts
var example2 string

func TestTS(t *testing.T) {
	vm := jz.New()
	_, err := vm.RunScript("example2.ts", example2)
	if err != nil {
		t.Error(err)
	}
}

func TestJS2x2(t *testing.T) {
	vm := jz.New()
	v, err := vm.RunString("2 + 2")
	if err != nil {
		t.Error(err)
		return
	}
	const want = 4
	if got := v.Export().(int64); got != want {
		t.Errorf("want %d, got %d", want, got)
	}
}

type MyStruct struct {
	Field int
}

func TestFieldMapper(t *testing.T) {
	vm := jz.New()
	ms := &MyStruct{Field: 123}
	err := vm.Set("ms", ms)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = vm.RunString("ms.field = 456")
	if err != nil {
		t.Error(err)
		return
	}
	const want = 456
	if ms.Field != want {
		t.Errorf("want %d, got %d", want, ms.Field)
	}
}

func TestRunFile(t *testing.T) {
	vm := jz.New()
	vm.ImportMap["three"] = "https://threejs.org/build/three.js"
	vm.ImportMap["three/addons/"] = "https://threejs.org/examples/jsm/"
	_, err := vm.RunFile("cmd/example1/example3.js")
	if err != nil {
		t.Error(err)
	}
}
