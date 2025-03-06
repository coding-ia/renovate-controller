package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/coding-ia/renovate-controller/internal/service"
	"github.com/google/go-github/v63/github"
	"os"
	"strconv"
	"strings"
)

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	applicationID := os.Getenv("GITHUB_APPLICATION_ID")
	clusterName := os.Getenv("AWS_ECS_CLUSTER_NAME")
	task := os.Getenv("AWS_ECS_CLUSTER_TASK")
	webhookSecret := os.Getenv("WEBHOOK_SECRET")
	assignPublicIP := os.Getenv("AWS_ECS_TASK_PUBLIC_IP")
	subnets := os.Getenv("AWS_ECS_TASK_SUBNET_IDS")
	securityGroups := os.Getenv("AWS_ECS_TASK_SECURITY_GROUP_IDS")

	publicIP, err := strconv.ParseBool(assignPublicIP)
	if err != nil {
		fmt.Println("Error parsing value for AWS_ECS_TASK_PUBLIC_IP")
	}

	eventType := request.Headers["X-GitHub-Event"]
	signature := request.Headers["X-Hub-Signature-256"]

	valid, err := validateGitHubSignature(request.Body, signature, []byte(webhookSecret))
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("{\"error\": \"Error validating signature: %s\"}", err),
			StatusCode: 200,
		}, nil
	}

	if !valid {
		fmt.Println("Invalid signature")
		return events.APIGatewayProxyResponse{
			Body:       "{\"error\": \"Invalid signature\"}",
			StatusCode: 200,
		}, nil
	}

	event, err := github.ParseWebHook(eventType, []byte(request.Body))
	if err != nil {
		fmt.Printf("Error parsing webhook: %s\n", err)
		return events.APIGatewayProxyResponse{
			Body:       "",
			StatusCode: 200,
		}, nil
	}

	var subnetsSlice []string
	var securityGroupsSlice []string

	if subnets != "" {
		subnetsSlice = strings.Split(subnets, ",")
	}
	if securityGroups != "" {
		securityGroupsSlice = strings.Split(securityGroups, ",")
	}

	config := service.ECSConfig{
		Cluster: clusterName,
		Task:    task,
		AWSVPCConfig: service.ECSVPCConfig{
			Subnets:        subnetsSlice,
			SecurityGroups: securityGroupsSlice,
			AssignPublicIP: publicIP,
		},
	}

	switch e := event.(type) {
	case *github.InstallationEvent:
		svc := service.NewRenovateTaskService(config)
		installationID := strconv.FormatInt(*e.Installation.ID, 10)

		taskConfig := service.RunTaskConfig{
			ApplicationID:  applicationID,
			InstallationID: installationID,
		}

		runTask(ctx, svc, taskConfig)
	case *github.InstallationRepositoriesEvent:
		svc := service.NewRenovateTaskService(config)
		installationID := strconv.FormatInt(*e.Installation.ID, 10)

		taskConfig := service.RunTaskConfig{
			ApplicationID:  applicationID,
			InstallationID: installationID,
		}

		runTask(ctx, svc, taskConfig)
	default:
		fmt.Printf("Unhandled event type: %s\n", eventType)
	}

	return events.APIGatewayProxyResponse{
		Body:       "",
		StatusCode: 200,
	}, nil
}

func runTask(ctx context.Context, svc *service.TaskService, taskConfig service.RunTaskConfig) {
	_, err := svc.RunTask(ctx, taskConfig)
	if err != nil {
		fmt.Printf("Error running task: %s\n", err)
	}
}

func validateGitHubSignature(body string, signature string, secret []byte) (bool, error) {
	if signature == "" {
		return false, fmt.Errorf("missing X-Hub-Signature-256 header")
	}

	// Ensure the signature starts with "sha256="
	if !strings.HasPrefix(signature, "sha256=") {
		return false, fmt.Errorf("invalid X-Hub-Signature-256 header format")
	}

	// Extract the signature value (without the "sha256=" prefix)
	expectedSignature := signature[7:]

	// Compute the HMAC-SHA256 hash of the body using the secret
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(body))
	computedSignature := hex.EncodeToString(mac.Sum(nil))

	// Compare the computed signature with the expected signature
	if hmac.Equal([]byte(computedSignature), []byte(expectedSignature)) {
		return true, nil
	}

	return false, nil
}

func main() {
	lambda.Start(handler)
}
