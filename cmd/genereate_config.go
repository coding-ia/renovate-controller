package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"renovate-controller/internal/processor"
)

var generateConfigCmd = &cobra.Command{
	Use:   "generate-config",
	Short: "Genereate renovate config",
	Long:  ``,
	Run:   generateConfigCommand,
}

func generateConfigCommand(cmd *cobra.Command, args []string) {
	appId := viper.GetString("appId")
	pemSecretArn := viper.GetString("pem-aws-secret")
	githubEndpoint := viper.GetString("endpoint")
	installationId := viper.GetInt64("installationId")
	targetRepository := viper.GetString("target-repository")
	s3Bucket := viper.GetString("s3-bucket")
	s3ConfigKey := viper.GetString("s3-config-key")
	output := viper.GetString("output")

	privateKey, err := parsePrivateKey(pemSecretArn)
	if err != nil {
		fmt.Printf("Error retrieving private key: %v\n", err)
		return
	}

	githubConfig := &processor.GitHubConfig{
		ApplicationID: appId,
		PrivateKey:    privateKey,
		Endpoint:      githubEndpoint,
	}

	options := processor.GenerateCommandOptions{
		InstallationID:   installationId,
		TargetRepository: targetRepository,
		Output:           output,
		S3Bucket:         s3Bucket,
		S3ConfigKey:      s3ConfigKey,
	}

	err = processor.Generate(githubConfig, options)
	if err != nil {
		log.Fatal(err)
	}
}
