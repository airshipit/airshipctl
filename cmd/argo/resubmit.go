package argo

import (
	"os"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/argoproj/argo/workflow/util"
	"github.com/argoproj/pkg/errors"

	env "github.com/ian-howell/airshipctl/pkg/environment"
)

func NewResubmitCommand() *cobra.Command {
	var (
		memoized      bool
		cliSubmitOpts cliSubmitOpts
	)
	var command = &cobra.Command{
		Use:   "resubmit WORKFLOW",
		Short: "resubmit a workflow",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}

			wfClient := InitWorkflowClient()
			wf, err := wfClient.Get(args[0], metav1.GetOptions{})
			errors.CheckError(err)
			newWF, err := util.FormulateResubmitWorkflow(wf, memoized)
			errors.CheckError(err)
			created, err := util.SubmitWorkflow(wfClient, newWF, nil)
			errors.CheckError(err)
			printWorkflow(created, cliSubmitOpts.output, env.Default)
			waitOrWatch([]string{created.Name}, cliSubmitOpts)
		},
	}

	command.Flags().StringVarP(&cliSubmitOpts.output, "output", "o", "", "Output format. One of: name|json|yaml|wide")
	command.Flags().BoolVarP(&cliSubmitOpts.wait, "wait", "w", false, "wait for the workflow to complete")
	command.Flags().BoolVar(&cliSubmitOpts.watch, "watch", false, "watch the workflow until it completes")
	command.Flags().BoolVar(&memoized, "memoized", false, "re-use successful steps & outputs from the previous run (experimental)")
	return command
}
