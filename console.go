package jz

import (
	"fmt"
	"log"
)

type Console struct {
	vm *Runtime
}

func NewConsole(vm *Runtime) Value {
	console := &Console{
		vm: vm,
	}
	return vm.ToValue(console)
}

func (console *Console) Log(args ...interface{}) {
	fmt.Println(args...)
}

func (console *Console) Info(args ...interface{}) {
	fmt.Println(args...)
}

func (console *Console) Error(args ...interface{}) {
	log.Println(args...)
}

func (console *Console) Warn(args ...interface{}) {
	log.Println(args...)
}
