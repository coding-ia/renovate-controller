package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run renovate",
	Long:  `Run renovate tasks in ECS cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Not yet enabled")
	},
}
