package main

import (
	"os"

	"github.com/ian-howell/airshipctl/cmd"
)

func main() {
	cmd.Execute(os.Stdout)
}
