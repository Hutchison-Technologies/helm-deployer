package filesystem

import (
	"os"
)

func IsFile(path string) bool {
	src, err := os.Stat(path)
	return !os.IsNotExist(err) && !src.IsDir()
}

func IsDirectory(path string) bool {
	src, err := os.Stat(path)
	return !os.IsNotExist(err) && !src.Mode().IsRegular()
}
