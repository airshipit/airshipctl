package main

import (
	"os"

	"github.com/ian-howell/airshipctl/cmd"
)

func main() {
	if err := RestoreAssets("", "_plugins"); err != nil {
		panic(err.Error())
	}
	cmd.Execute(os.Stdout)
}
