package controller

import (
	"github.com/coding-ia/renovate-controller/cmd"
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
	cmd.taskCmd.PersistentFlags().StringP("appId", "a", "", "GitHub Installation Application ID")
	cmd.taskCmd.PersistentFlags().StringP("pem-aws-secret", "s", "", "GitHub Application Private Key (Secrets Manager)")
	cmd.taskCmd.PersistentFlags().StringP("endpoint", "e", "", "GitHub Endpoint")

	mapEnvToPFlag(cmd.taskCmd, "appId", "GITHUB_APPLICATION_ID")
	mapEnvToPFlag(cmd.taskCmd, "pem-aws-secret", "GITHUB_APPLICATION_PRIVATE_PEM_AWS_SECRET")
	mapEnvToPFlag(cmd.taskCmd, "endpoint", "GITHUB_APPLICATION_ENDPOINT")

	cmd.runCmd.Flags().StringP("cluster", "c", "", "ECS Cluster Name")
	cmd.runCmd.Flags().StringP("task", "t", "", "Task Definition Name")
	cmd.runCmd.Flags().String("container-name", "renovate", "Task Container Name")
	cmd.runCmd.Flags().String("subnet-ids", "", "AWS VPC Subnet IDs")
	cmd.runCmd.Flags().String("security-group-ids", "", "AWS VPC SecurityGroup IDs")
	cmd.runCmd.Flags().Bool("assign-public-ip", false, "Assign Public IP to Task")

	mapEnvToFlag(cmd.runCmd, "cluster", "AWS_ECS_CLUSTER_NAME")
	mapEnvToFlag(cmd.runCmd, "task", "AWS_ECS_CLUSTER_TASK")
	mapEnvToFlag(cmd.runCmd, "container-name", "AWS_ECS_CLUSTER_TASK_CONTAINER_NAME")
	mapEnvToFlag(cmd.runCmd, "subnet-ids", "AWS_ECS_TASK_SUBNET_IDS")
	mapEnvToFlag(cmd.runCmd, "security-group-ids", "AWS_ECS_TASK_SECURITY_GROUP_IDS")
	mapEnvToFlag(cmd.runCmd, "assign-public-ip", "AWS_ECS_TASK_PUBLIC_IP")

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

	cmd.taskCmd.AddCommand(cmd.runCmd)
	cmd.taskCmd.AddCommand(generateConfigCmd)
	rootCmd.AddCommand(cmd.taskCmd)

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
