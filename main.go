package main

import (
	"fmt"
	"os"

	"github.com/ian-howell/airshipctl/cmd"
	"github.com/ian-howell/airshipctl/pkg/log"
)

func main() {
	rootCmd, settings, err := cmd.NewRootCmd(os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}

	cmd.AddDefaultAirshipCTLCommands(rootCmd, settings)

	// Flags may not be parsed until all subcommands have been added
	rootCmd.PersistentFlags().Parse(os.Args[1:])

	settings.Init()

	log.Init(settings, os.Stdout)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}
