package main

import (
	"fmt"
	"github.com/andrylavr/jz"
	"log"
)
import _ "embed"

//go:embed example1.js
var example1 string

//go:embed example2.ts
var example2 string

func main() {
	runExample1JS()
	runExample2TS()
	runExample3JS()
	runExample4FieldMapper()
}

func notOk(err error) bool {
	if err != nil {
		log.Println(err)
		return true
	}
	return false
}

func runExample1JS() {
	vm := jz.New()
	vm.ImportMap["three"] = "https://threejs.org/build/three.js"
	vm.ImportMap["three/addons/"] = "https://threejs.org/examples/jsm/"

	_, err := vm.RunScript("example1.js", example1)
	if notOk(err) {
		return
	}
}

func runExample2TS() {
	vm := jz.New()
	_, err := vm.RunScript("example2.ts", example2)
	if err != nil {
		log.Println(err)
	}
}

func runExample3JS() {
	vm := jz.New()
	v, err := vm.RunString("2 + 2")
	if err != nil {
		panic(err)
	}
	if num := v.Export().(int64); num != 4 {
		panic(num)
	}
}

type MyStruct struct {
	Field int
}

func (s *MyStruct) Test(f float64) {
	fmt.Println("MyStruct.Test", s.Field, f)
}

func runExample4FieldMapper() {
	vm := jz.New()
	err := vm.Set("ms", &MyStruct{Field: 123})
	if notOk(err) {
		return
	}
	_, err = vm.RunString("ms.test(456.7)")
	if notOk(err) {
		return
	}
}
