package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	"renovate-controller/internal/aws_helper"
	"renovate-controller/internal/processor"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run renovate",
	Long:  `Run renovate tasks in ECS cluster`,
	Run:   runCommand,
}

func init() {
	runCmd.Flags().StringP("appId", "a", "", "GitHub Application ID")
	runCmd.Flags().StringP("pem", "p", "", "GitHub Application Private Key File")
	runCmd.Flags().StringP("pem-aws-secret", "s", "", "GitHub Application Private Key (Secrets Manager)")
	runCmd.Flags().StringP("endpoint", "e", "", "GitHub Endpoint")
	runCmd.Flags().StringP("cluster", "c", "", "ECS Cluster Name")
	runCmd.Flags().StringP("task", "t", "", "Task Definition Name")

	mapEnvToFlag(runCmd, "appId", "GITHUB_APP_ID")
	mapEnvToFlag(runCmd, "pem", "GITHUB_APP_PRIVATE_KEY_FILE")
	mapEnvToFlag(runCmd, "pem-aws-secret", "GITHUB_APP_PRIVATE_KEY_AWS_SECRET")
	mapEnvToFlag(runCmd, "endpoint", "GITHUB_ENDPOINT")
	mapEnvToFlag(runCmd, "cluster", "AWS_ECS_CLUSTER_NAME")
	mapEnvToFlag(runCmd, "task", "AWS_ECS_CLUSTER_TASK")

	rootCmd.AddCommand(runCmd)
}

func runCommand(cmd *cobra.Command, args []string) {
	appId := viper.GetString("appId")
	pemFile := viper.GetString("pem")
	pemSecretArn := viper.GetString("pem-aws-secret")
	githubEndpoint := viper.GetString("endpoint")

	privateKey, err := parsePrivateKey(pemFile, pemSecretArn)
	if err != nil {
		fmt.Printf("Error retrieving private key: %v\n", err)
		return
	}

	task := viper.GetString("task")
	clusterName := viper.GetString("cluster")

	config := &processor.RenovateTaskConfig{
		TaskDefinition: task,
		ClusterName:    clusterName,
		AssignPublicIP: true,
	}

	err = processor.RunRenovateTasks(appId, privateKey, githubEndpoint, config)
	if err != nil {
		fmt.Printf("Error running renovate: %v\n", err)
		return
	}
}

func parsePrivateKey(pemFile string, pemSecretArn string) ([]byte, error) {
	if pemSecretArn != "" {
		secret, err := aws_helper.GetSecret(pemSecretArn)
		if err != nil {
			return nil, err
		}
		return []byte(secret), nil
	}

	privateKey, err := os.ReadFile(pemFile)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func mapEnvToFlag(command *cobra.Command, flag string, env string) {
	err := viper.BindPFlag(flag, command.Flags().Lookup(flag))
	if err != nil {
		log.Fatal(err)
	}
	err = viper.BindEnv(flag, env)
	if err != nil {
		log.Fatalln(err)
	}
}
