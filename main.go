package main

import (
	"github.com/Hutchison-Technologies/helm-deployer/cli"
)

func main() {
	err := cli.Run()
	if err != nil {
		panic(err.Error())
	}
}
