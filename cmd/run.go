package cmd

import (
	"fmt"
	"github.com/coding-ia/renovate-controller/internal/processor"
	"github.com/coding-ia/renovate-controller/internal/secrets"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"strings"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run renovate",
	Long:  `Run renovate tasks in ECS cluster`,
	Run:   runCommand,
}

func runCommand(cmd *cobra.Command, args []string) {
	appId := viper.GetString("appId")
	pemSecretArn := viper.GetString("pem-aws-secret")
	githubEndpoint := viper.GetString("endpoint")
	containerName := viper.GetString("container-name")
	subnets := viper.GetString("subnet-ids")
	securityGroups := viper.GetString("security-group-ids")
	publicIP := viper.GetBool("assign-public-ip")

	privateKey, err := parsePrivateKey(pemSecretArn)
	if err != nil {
		fmt.Printf("Error retrieving private key: %v\n", err)
		return
	}

	task := viper.GetString("task")
	clusterName := viper.GetString("cluster")

	var subnetsSlice []string
	var securityGroupsSlice []string

	if subnets != "" {
		subnetsSlice = strings.Split(subnets, ",")
	}
	if securityGroups != "" {
		securityGroupsSlice = strings.Split(securityGroups, ",")
	}

	runConfig := &processor.RunCommandOptions{
		TaskDefinition: task,
		ClusterName:    clusterName,
		ContainerName:  containerName,
		AssignPublicIP: publicIP,
		Subnets:        subnetsSlice,
		SecurityGroups: securityGroupsSlice,
		TaskOptions: processor.TaskCommandOptions{
			ApplicationID: appId,
			PEMAWSSecret:  pemSecretArn,
			Endpoint:      githubEndpoint,
		},
	}

	githubConfig := &processor.GitHubConfig{
		ApplicationID: appId,
		PrivateKey:    privateKey,
		Endpoint:      githubEndpoint,
	}

	err = processor.Run(githubConfig, runConfig)
	if err != nil {
		log.Fatal(err)
	}
}

func parsePrivateKey(pemSecretArn string) ([]byte, error) {
	secret, err := secrets.GetSecret(pemSecretArn)
	if err != nil {
		return nil, err
	}
	return []byte(secret), nil
}
