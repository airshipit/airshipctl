package main

import (
	"fmt"
	"os"

	"github.com/ian-howell/airshipctl/cmd"
	"github.com/ian-howell/airshipctl/cmd/workflow"

	"github.com/ian-howell/airshipctl/pkg/environment"
	"github.com/ian-howell/airshipctl/pkg/log"
)

func main() {
	rootCmd, err := cmd.NewRootCmd(os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}

	settings := &environment.AirshipCTLSettings{}
	settings.InitFlags(rootCmd)

	rootCmd.AddCommand(workflow.NewWorkflowCommand(os.Stdout, settings))

	rootCmd.PersistentFlags().Parse(os.Args[1:])

	settings.Init()

	log.Init(settings, os.Stdout)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}
