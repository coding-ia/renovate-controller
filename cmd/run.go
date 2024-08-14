package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"renovate-controller/internal/aws_helper"
	"renovate-controller/internal/processor"
)

var clusterName string
var taskDefinition string

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run renovate",
	Long:  `Run renovate tasks in ECS cluster`,
	Run:   runCommand,
}

func init() {
	runCmd.Flags().StringVarP(&applicationId, "appId", "a", "", "GitHub Application ID")
	runCmd.Flags().StringVarP(&applicationPEM, "pem", "p", "", "GitHub Application Private Key File")
	runCmd.Flags().StringVarP(&applicationPEMSecret, "pem-aws-secret", "s", "", "GitHub Application Private Key (Secrets Manager)")
	runCmd.Flags().StringVarP(&endpoint, "endpoint", "e", "", "GitHub Endpoint")

	runCmd.Flags().StringVarP(&clusterName, "cluster", "c", "", "ECS Cluster Name")
	runCmd.Flags().StringVarP(&taskDefinition, "task", "t", "", "Task Definition Name")
}

func runCommand(cmd *cobra.Command, args []string) {
	privateKey, err := parsePrivateKey()
	if err != nil {
		fmt.Printf("Error retrieving private key: %v\n", err)
		return
	}

	config := &processor.RenovateTaskConfig{
		TaskDefinition: taskDefinition,
		ClusterName:    clusterName,
		AssignPublicIP: true,
	}

	err = processor.RunRenovateTasks(applicationId, privateKey, endpoint, config)
	if err != nil {
		fmt.Printf("Error running renovate: %v\n", err)
		return
	}
}

func parsePrivateKey() ([]byte, error) {
	if applicationPEMSecret != "" {
		secret, err := aws_helper.GetSecret(applicationPEMSecret)
		if err != nil {
			return nil, err
		}
		return []byte(secret), nil
	}

	privateKey, err := os.ReadFile(applicationPEM)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}
