package main

import (
	"os"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func IsDirectory(path string) bool {
	src, err := os.Stat(path)

	return !os.IsNotExist(err) && !src.Mode().IsRegular()
}
