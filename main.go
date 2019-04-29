package main

import (
	"os"

	"github.com/ian-howell/airshipadm/cmd"
)

func main() {
	cmd.Execute(os.Stdout)
}
