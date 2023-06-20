package jz

import (
	"io"
	"net/http"
	"os"
	"strings"
)

func GetContent(url string) (string, error) {
	if strings.Contains(url, "http://") || strings.Contains(url, "https://") {
		resp, err := http.Get(url)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return string(b), err
	}

	b, err := os.ReadFile(url)
	if err != nil {
		return "", err
	}
	return string(b), err
}
