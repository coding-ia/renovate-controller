package service

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"log"
)

type ECSConfig struct {
	Cluster   string
	Task      string
	Container string
	PublicIP  bool
}

type TaskService struct {
	Config ECSConfig
}

type RenovateTaskService interface {
	RunTask(installationToken string, repository string, endpoint string) (*ecs.RunTaskOutput, error)
}

func NewRenovateTaskService(config ECSConfig) *TaskService {
	return &TaskService{
		Config: config,
	}
}

func (t *TaskService) RunTask(installationToken string, repository string, endpoint string) (*ecs.RunTaskOutput, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	svc := ecs.NewFromConfig(cfg)

	subnets, err := filterSubnets()
	if err != nil {
		return nil, err
	}
	if subnets == nil {
		return nil, fmt.Errorf("no subnets found")
	}

	assignPublicIP := types.AssignPublicIpDisabled
	securityGroups, err := filterSecurityGroups()
	if t.Config.PublicIP {
		assignPublicIP = types.AssignPublicIpEnabled
	} else {
		if securityGroups == nil {
			log.Printf("No security groups found and public IP disabled.")
		}
	}

	runTaskInput := &ecs.RunTaskInput{
		Cluster:        aws.String(t.Config.Cluster),
		TaskDefinition: aws.String(t.Config.Task),
		LaunchType:     types.LaunchTypeFargate,
		NetworkConfiguration: &types.NetworkConfiguration{
			AwsvpcConfiguration: &types.AwsVpcConfiguration{
				AssignPublicIp: assignPublicIP,
				Subnets:        subnets,
				SecurityGroups: securityGroups,
			},
		},
		Overrides: &types.TaskOverride{
			ContainerOverrides: []types.ContainerOverride{
				{
					Name: aws.String(t.Config.Container),
					Environment: []types.KeyValuePair{
						{
							Name:  aws.String("RENOVATE_ENDPOINT"),
							Value: aws.String(endpoint),
						},
						{
							Name:  aws.String("RENOVATE_TOKEN"),
							Value: aws.String(installationToken),
						},
						{
							Name:  aws.String("RENOVATE_REPOSITORIES"),
							Value: aws.String(repository),
						},
					},
				},
			},
		},
	}

	runTaskOutput, err := svc.RunTask(context.TODO(), runTaskInput)
	if err != nil {
		return nil, err
	}

	return runTaskOutput, nil
}

func filterSubnets() ([]string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	ec2Client := ec2.NewFromConfig(cfg)

	describeSubnetsInput := &ec2.DescribeSubnetsInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("tag:allow-renovate"),
				Values: []string{"true"},
			},
		},
	}

	result, err := ec2Client.DescribeSubnets(context.TODO(), describeSubnetsInput)
	if err != nil {
		return nil, err
	}

	var subnetIDs []string
	for _, subnet := range result.Subnets {
		subnetIDs = append(subnetIDs, *subnet.SubnetId)
	}

	return subnetIDs, nil
}

func filterSecurityGroups() ([]string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	ec2Client := ec2.NewFromConfig(cfg)

	filters := []ec2types.Filter{
		{
			Name:   aws.String("tag:renovate"),
			Values: []string{"true"},
		},
	}

	result, err := ec2Client.DescribeSecurityGroups(context.TODO(), &ec2.DescribeSecurityGroupsInput{
		Filters: filters,
	})
	if err != nil {
		return nil, err
	}

	var securityGroupIDs []string
	for _, sg := range result.SecurityGroups {
		securityGroupIDs = append(securityGroupIDs, *sg.GroupId)
	}

	return securityGroupIDs, nil
}
