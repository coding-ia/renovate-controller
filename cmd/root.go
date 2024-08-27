package cmd

import (
	"github.com/spf13/viper"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "renovate-controller",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

func Execute() {
	taskCmd.PersistentFlags().StringP("appId", "a", "", "GitHub Installation Application ID")
	taskCmd.PersistentFlags().StringP("pem-aws-secret", "s", "", "GitHub Application Private Key (Secrets Manager)")
	taskCmd.PersistentFlags().StringP("endpoint", "e", "", "GitHub Endpoint")

	mapEnvToPFlag(taskCmd, "appId", "GITHUB_APPLICATION_ID")
	mapEnvToPFlag(taskCmd, "pem-aws-secret", "GITHUB_APPLICATION_PRIVATE_PEM_AWS_SECRET")
	mapEnvToPFlag(taskCmd, "endpoint", "GITHUB_APPLICATION_ENDPOINT")

	runCmd.Flags().StringP("cluster", "c", "", "ECS Cluster Name")
	runCmd.Flags().StringP("task", "t", "", "Task Definition Name")
	runCmd.Flags().String("container-name", "renovate", "Task Container Name")
	runCmd.Flags().String("subnet-ids", "", "AWS VPC Subnet IDs")
	runCmd.Flags().String("security-group-ids", "", "AWS VPC SecurityGroup IDs")
	runCmd.Flags().Bool("assign-public-ip", false, "Assign Public IP to Task")

	mapEnvToFlag(runCmd, "cluster", "AWS_ECS_CLUSTER_NAME")
	mapEnvToFlag(runCmd, "task", "AWS_ECS_CLUSTER_TASK")
	mapEnvToFlag(runCmd, "container-name", "AWS_ECS_CLUSTER_TASK_CONTAINER_NAME")
	mapEnvToFlag(runCmd, "subnet-ids", "AWS_ECS_TASK_SUBNET_IDS")
	mapEnvToFlag(runCmd, "security-group-ids", "AWS_ECS_TASK_SECURITY_GROUP_IDS")
	mapEnvToFlag(runCmd, "assign-public-ip", "AWS_ECS_TASK_PUBLIC_IP")

	generateConfigCmd.Flags().Int64P("installationId", "", 0, "GitHub Installation ID")
	generateConfigCmd.Flags().StringP("target-repository", "", "", "GitHub target repository")
	generateConfigCmd.Flags().StringP("s3-bucket", "", "", "Renovate config (AWS S3 Bucket)")
	generateConfigCmd.Flags().StringP("s3-config-key", "", "", "Renovate config file (AWS S3 Bucket Key)")
	generateConfigCmd.Flags().StringP("output", "o", "config.ts", "Config file")

	mapEnvToFlag(generateConfigCmd, "installationId", "GITHUB_INSTALLATION_ID")
	mapEnvToFlag(generateConfigCmd, "target-repository", "GITHUB_TARGET_REPOSITORY")
	mapEnvToFlag(generateConfigCmd, "s3-bucket", "CONFIG_TEMPLATE_BUCKET")
	mapEnvToFlag(generateConfigCmd, "s3-config-key", "CONFIG_TEMPLATE_KEY")
	mapEnvToFlag(generateConfigCmd, "output", "GENERATE_CONFIG_OUTPUT")

	taskCmd.AddCommand(runCmd)
	taskCmd.AddCommand(generateConfigCmd)
	rootCmd.AddCommand(taskCmd)

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
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
func mapEnvToPFlag(command *cobra.Command, flag string, env string) {
	err := viper.BindPFlag(flag, command.PersistentFlags().Lookup(flag))
	if err != nil {
		log.Fatal(err)
	}
	err = viper.BindEnv(flag, env)
	if err != nil {
		log.Fatalln(err)
	}
}
