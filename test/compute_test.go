package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"test/utils"
)

func TestComputeModule(t *testing.T) {
	t.Parallel()

	workingDir := test_structure.CopyTerraformFolderToTemp(t, "../", "modules/compute")
	uniqueID := random.UniqueId()[:6]
	projectName := fmt.Sprintf("comp%s", uniqueID)

	testCases := []struct {
		name        string
		region      string
		environment string
	}{
		{
			name:        "us-east-1-ci",
			region:      "us-east-1",
			environment: "ci",
		},
	}

	for _, testCase := range testCases {
		tc := testCase

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			test_structure.RunTestStage(t, "setup", func() {
				setupComputeTest(t, workingDir, tc.region, tc.environment, projectName)
			})

			test_structure.RunTestStage(t, "validate", func() {
				validateComputeInfrastructure(t, workingDir)
			})
		})
	}
}

func setupComputeTest(t *testing.T, workingDir, region, environment, projectName string) {
	// Create VPC first as compute depends on it
	vpcOpts := utils.CreateVPC(t, region, environment, projectName)
	defer terraform.Destroy(t, vpcOpts)

	terraform.InitAndApply(t, vpcOpts)

	vpcID := terraform.Output(t, vpcOpts, "vpc_id")
	privateSubnets := terraform.OutputList(t, vpcOpts, "private_subnets")
	publicSubnets := terraform.OutputList(t, vpcOpts, "public_subnets")

	// Create ALB using the ALB module
	albOpts := utils.CreateALB(t, region, environment, projectName, vpcID, publicSubnets)
	defer terraform.Destroy(t, albOpts)

	terraform.InitAndApply(t, albOpts)

	// Get the outputs from the ALB module
	albSecurityGroupID := terraform.Output(t, albOpts, "security_group_id")

	// Get target group ARNs from the ALB module output
	targetGroupArns := []string{
		terraform.Output(t, albOpts, "target_group_arns.app1"),
		terraform.Output(t, albOpts, "target_group_arns.app2"),
	}

	// Setup compute module
	computeOpts := &terraform.Options{
		TerraformDir: workingDir,
		Vars: map[string]interface{}{
			"environment":           environment,
			"project_name":          projectName,
			"vpc_id":                vpcID,
			"private_subnets":       privateSubnets,
			"instance_type":         "t3.micro",
			"instance_count":        2,
			"apps":                  albOpts.Vars["apps"],
			"target_group_arns":     targetGroupArns,
			"alb_security_group_id": albSecurityGroupID,
		},
		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": region,
		},
	}

	test_structure.SaveTerraformOptions(t, workingDir, computeOpts)

	terraform.InitAndApply(t, computeOpts)
}

func validateComputeInfrastructure(t *testing.T, workingDir string) {
	computeOpts := test_structure.LoadTerraformOptions(t, workingDir)
	awsRegion := computeOpts.EnvVars["AWS_DEFAULT_REGION"]

	ec2Client := utils.CreateEC2Client(awsRegion)
	asgClient := utils.CreateASGClient(awsRegion)
	iamClient := utils.CreateIAMClient(awsRegion)

	// Test Launch Template
	testLaunchTemplate(t, ec2Client, computeOpts)

	// Test Auto Scaling Group
	testAutoScalingGroup(t, asgClient, computeOpts)

	// Test IAM Role and Instance Profile
	testIAMConfiguration(t, iamClient, computeOpts)

	// Test Security Group
	testSecurityGroup(t, ec2Client, computeOpts)
}

func testLaunchTemplate(t *testing.T, ec2Client *ec2.EC2, terraformOptions *terraform.Options) {
	ltID := terraform.Output(t, terraformOptions, "launch_template_id")

	input := &ec2.DescribeLaunchTemplatesInput{
		LaunchTemplateIds: []*string{aws.String(ltID)},
	}

	result, err := ec2Client.DescribeLaunchTemplates(input)
	require.NoError(t, err)
	require.Len(t, result.LaunchTemplates, 1)

	// Get latest version details
	versionInput := &ec2.DescribeLaunchTemplateVersionsInput{
		LaunchTemplateId: aws.String(ltID),
		Versions:         []*string{aws.String("$Latest")},
	}

	versionResult, err := ec2Client.DescribeLaunchTemplateVersions(versionInput)
	require.NoError(t, err)
	require.Len(t, versionResult.LaunchTemplateVersions, 1)

	lt := versionResult.LaunchTemplateVersions[0]

	// Verify instance type
	assert.Equal(t, terraformOptions.Vars["instance_type"], *lt.LaunchTemplateData.InstanceType)

	// Verify EBS volume
	require.Len(t, lt.LaunchTemplateData.BlockDeviceMappings, 1)
	volume := lt.LaunchTemplateData.BlockDeviceMappings[0].Ebs
	assert.Equal(t, int64(30), *volume.VolumeSize)
	assert.Equal(t, "gp3", *volume.VolumeType)

	// Verify tags
	hasExpectedTags := utils.HasRequiredTags(
		utils.ConvertEC2TagsToTags(lt.LaunchTemplateData.TagSpecifications[0].Tags),
		map[string]string{
			"Environment": terraformOptions.Vars["environment"].(string),
			"Project":     terraformOptions.Vars["project_name"].(string),
		},
	)
	assert.True(t, hasExpectedTags)
}

func testAutoScalingGroup(t *testing.T, asgClient *autoscaling.AutoScaling, terraformOptions *terraform.Options) {
	asgName := terraform.Output(t, terraformOptions, "autoscaling_group_name")

	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String(asgName)},
	}

	result, err := asgClient.DescribeAutoScalingGroups(input)
	require.NoError(t, err)
	require.Len(t, result.AutoScalingGroups, 1)

	asg := result.AutoScalingGroups[0]

	// Verify instance count
	assert.Equal(t, int64(terraformOptions.Vars["instance_count"].(int)), *asg.DesiredCapacity)
	assert.Equal(t, int64(terraformOptions.Vars["instance_count"].(int)), *asg.MinSize)
	assert.Equal(t, int64(terraformOptions.Vars["instance_count"].(int)*2), *asg.MaxSize)

	// Verify subnets
	expectedSubnets := terraformOptions.Vars["private_subnets"].([]string)
	actualSubnets := strings.Split(*asg.VPCZoneIdentifier, ",")
	assert.ElementsMatch(t, expectedSubnets, actualSubnets)

	// Verify target groups
	expectedTargetGroups := terraformOptions.Vars["target_group_arns"].([]string)
	assert.ElementsMatch(t, expectedTargetGroups, aws.StringValueSlice(asg.TargetGroupARNs))
}

func testIAMConfiguration(t *testing.T, iamClient *iam.IAM, terraformOptions *terraform.Options) {
	roleName := terraform.Output(t, terraformOptions, "iam_role_name")

	// Check role
	roleInput := &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	}

	roleResult, err := iamClient.GetRole(roleInput)
	require.NoError(t, err)

	// Verify role trust policy
	assert.Contains(t, *roleResult.Role.AssumeRolePolicyDocument, "ec2.amazonaws.com")

	// Check attached policies
	policiesInput := &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	}

	policiesResult, err := iamClient.ListAttachedRolePolicies(policiesInput)
	require.NoError(t, err)

	// Verify S3 read access policy is attached
	foundS3Policy := false
	for _, policy := range policiesResult.AttachedPolicies {
		if *policy.PolicyArn == "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess" {
			foundS3Policy = true
			break
		}
	}
	assert.True(t, foundS3Policy, "S3 read only policy should be attached to the role")
}

func testSecurityGroup(t *testing.T, ec2Client *ec2.EC2, terraformOptions *terraform.Options) {
	sgID := terraform.Output(t, terraformOptions, "security_group_id")

	input := &ec2.DescribeSecurityGroupsInput{
		GroupIds: []*string{aws.String(sgID)},
	}

	result, err := ec2Client.DescribeSecurityGroups(input)
	require.NoError(t, err)
	require.Len(t, result.SecurityGroups, 1)

	sg := result.SecurityGroups[0]

	// Verify inbound rules (one for each app port)
	apps := terraformOptions.Vars["apps"].(map[string]interface{})
	assert.Len(t, sg.IpPermissions, len(apps))

	for _, rule := range sg.IpPermissions {
		port := *rule.FromPort
		foundMatchingApp := false
		for _, app := range apps {
			if app.(map[string]interface{})["port"].(int) == int(port) {
				foundMatchingApp = true
				break
			}
		}
		assert.True(t, foundMatchingApp, fmt.Sprintf("Found unexpected port %d in security group", port))
	}
}

func createMockALBSecurityGroup(t *testing.T, ec2Client *ec2.EC2, vpcID, projectName, environment string) string {
	input := &ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(fmt.Sprintf("%s-%s-mock-alb-sg", projectName, environment)),
		Description: aws.String("Mock ALB Security Group for testing"),
		VpcId:       aws.String(vpcID),
	}

	result, err := ec2Client.CreateSecurityGroup(input)
	require.NoError(t, err)

	// Add tags
	_, err = ec2Client.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{result.GroupId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(fmt.Sprintf("%s-%s-mock-alb-sg", projectName, environment)),
			},
		},
	})
	require.NoError(t, err)

	return *result.GroupId
}