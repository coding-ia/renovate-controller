package cmd

import "github.com/spf13/cobra"

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "",
	Long:  ``,
	Run:   taskCmdFunc,
}

func taskCmdFunc(cmd *cobra.Command, args []string) {
	_ = cmd.Help()
}
