package test

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/require"

	"test/utils"
)

func TestE2E(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		region      string
		environment string
		projectName string
	}{
		{
			name:        "Complete Example",
			region:      "us-east-1",
			environment: "test",
			projectName: "e2e",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			terraformOptions := &terraform.Options{
				TerraformDir: "../examples/complete",
				Vars: map[string]interface{}{
					"environment":  testCase.environment,
					"project_name": testCase.projectName,
				},
				EnvVars: map[string]string{
					"AWS_DEFAULT_REGION": testCase.region,
				},
			}

			defer terraform.Destroy(t, terraformOptions)

			terraform.InitAndApply(t, terraformOptions)

			// Create AWS clients
			ec2Client := utils.CreateEC2Client(testCase.region)
			asgClient := utils.CreateASGClient(testCase.region)
			//iamClient := utils.CreateIAMClient(testCase.region)

			// Get outputs
			vpcID := terraform.Output(t, terraformOptions, "vpc_id")
			launchTemplateID := terraform.Output(t, terraformOptions, "launch_template_id")
			asgName := terraform.Output(t, terraformOptions, "autoscaling_group_name")

			// Test VPC
			vpc, err := ec2Client.DescribeVpcs(&ec2.DescribeVpcsInput{
				VpcIds: []*string{&vpcID},
			})
			require.NoError(t, err)
			require.Len(t, vpc.Vpcs, 1)

			// Test Launch Template
			lt, err := ec2Client.DescribeLaunchTemplates(&ec2.DescribeLaunchTemplatesInput{
				LaunchTemplateIds: []*string{&launchTemplateID},
			})
			require.NoError(t, err)
			require.Len(t, lt.LaunchTemplates, 1)

			// Test Auto Scaling Group
			asg, err := asgClient.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
				AutoScalingGroupNames: []*string{&asgName},
			})
			require.NoError(t, err)
			require.Len(t, asg.AutoScalingGroups, 1)

			// Wait for instances to be running
			time.Sleep(2 * time.Minute)

			// Verify instances are running
			instances, err := ec2Client.DescribeInstances(&ec2.DescribeInstancesInput{
				Filters: []*ec2.Filter{
					{
						Name:   aws.String("vpc-id"),
						Values: []*string{&vpcID},
					},
					{
						Name:   aws.String("instance-state-name"),
						Values: []*string{aws.String("running")},
					},
				},
			})
			require.NoError(t, err)
			require.NotEmpty(t, instances.Reservations)
		})
	}
}
