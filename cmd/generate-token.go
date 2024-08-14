package cmd

import (
	"fmt"
	"os"
	"renovate-controller/internal/github_helper"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(generateCmd)
}

var applicationId string
var applicationPEM string
var endpoint string
var owner string
var repository string

var generateCmd = &cobra.Command{
	Use:   "generate-token",
	Short: "Generate GitHub installation token",
	Long:  `Generate GitHub installation token for renovate bot`,
	Run:   generateTokenCommand,
}

func init() {
	generateCmd.Flags().StringVarP(&applicationId, "appId", "a", "", "GitHub Application ID")
	generateCmd.Flags().StringVarP(&applicationPEM, "pem", "p", "", "GitHub Application Private Key File")
	generateCmd.Flags().StringVarP(&endpoint, "endpoint", "e", "", "GitHub Endpoint")
	generateCmd.Flags().StringVarP(&owner, "owner", "o", "", "GitHub Repository Owner")
	generateCmd.Flags().StringVarP(&repository, "repository", "r", "", "GitHub Repository")
}

func generateTokenCommand(cmd *cobra.Command, args []string) {
	privateKey, err := os.ReadFile(applicationPEM)
	if err != nil {
		fmt.Printf("Error reading private key: %v\n", err)
		return
	}

	token, err := github_helper.GenerateInstallationToken(applicationId, privateKey, endpoint, owner, repository)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Installation Token: %s\n", token)
}
