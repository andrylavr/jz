package jz

import (
	"crypto/md5"
	"errors"
	"fmt"
	"os"
)

type Cacher interface {
	Exists(key string) bool
	Get(key string) string
	Save(key string, s string)
}

type FSCacher struct {
	Dir string
}

func (t FSCacher) getFilename(key string) string {
	m := md5.Sum([]byte(key))
	s := fmt.Sprintf("%x", m)
	return t.Dir + "/" + s + ".js"
}

func (t FSCacher) Exists(key string) bool {
	filename := t.getFilename(key)
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func (t FSCacher) Get(key string) string {
	filename := t.getFilename(key)
	b, _ := os.ReadFile(filename)
	return string(b)
}

func (t FSCacher) Save(key string, content string) {
	filename := t.getFilename(key)
	os.WriteFile(filename, []byte(content), os.ModePerm)
}
