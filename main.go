package main

import (
	"github.com/Hutchison-Technologies/bluegreen-deployer/cli"
)

func main() {
	err := cli.Run()
	if err != nil {
		panic(err.Error())
	}
}
