package main

import (
	"fmt"
	"os"

	"github.com/ian-howell/airshipctl/cmd"
)

func main() {
	rootCmd, _, err := cmd.NewAirshipCTLCommand(os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}
