package test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	terratest_aws "github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to check if a map contains specific tags
func hasRequiredTags(tags []*ec2.Tag, requiredTags map[string]string) bool {
	tagMap := make(map[string]string)
	for _, tag := range tags {
		tagMap[*tag.Key] = *tag.Value
	}

	for key, value := range requiredTags {
		actualValue, exists := tagMap[key]
		if !exists {
			fmt.Printf("Missing required tag: %s\n", key)
			return false
		}
		if actualValue != value {
			fmt.Printf("Tag mismatch for %s: expected '%s', got '%s'\n", key, value, actualValue)
			return false
		}
	}
	return true
}

func TestVPCModule(t *testing.T) {
	t.Parallel()

	// Generate a random name to prevent a naming conflict
	uniqueID := random.UniqueId()
	projectName := fmt.Sprintf("vpc-test-%s", uniqueID)

	// Construct the test cases
	testCases := []struct {
		name        string
		region      string
		environment string
		vpcCidr     string
	}{
		{
			name:        "us-east-1-staging",
			region:      "us-east-1",
			environment: "staging",
			vpcCidr:     "10.0.0.0/16",
		},
		{
			name:        "eu-west-1-staging",
			region:      "eu-west-1",
			environment: "staging",
			vpcCidr:     "10.1.0.0/16",
		},
	}

	for _, testCase := range testCases {
		// Local scope for parallel tests
		tc := testCase

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			terraformOptions := &terraform.Options{
				// The path to where your Terraform code is located
				TerraformDir: "../modules/vpc",

				// Variables to pass to our Terraform code using -var options
				Vars: map[string]interface{}{
					"region":       tc.region,
					"environment":  tc.environment,
					"project_name": projectName,
					"vpc_cidr":     tc.vpcCidr,
				},

				// Environment variables to set when running Terraform
				EnvVars: map[string]string{
					"AWS_DEFAULT_REGION": tc.region,
				},
			}

			// At the end of the test, run `terraform destroy`
			defer terraform.Destroy(t, terraformOptions)

			// Run `terraform init` and `terraform apply`
			terraform.InitAndApply(t, terraformOptions)

			// Get VPC ID from Terraform outputs
			vpcID := terraform.Output(t, terraformOptions, "vpc_id")
			require.NotEmpty(t, vpcID, "VPC ID should not be empty")

			// Create AWS EC2 service client
			ec2Client := terratest_aws.NewEc2Client(t, tc.region)

			// Verify VPC exists and check CIDR
			vpcOutput, err := ec2Client.DescribeVpcs(&ec2.DescribeVpcsInput{
				VpcIds: []*string{aws.String(vpcID)},
			})
			require.NoError(t, err)
			require.Len(t, vpcOutput.Vpcs, 1)
			assert.Equal(t, tc.vpcCidr, *vpcOutput.Vpcs[0].CidrBlock)

			// Get subnet IDs from Terraform outputs
			privateSubnetsStr := terraform.OutputList(t, terraformOptions, "private_subnets")
			publicSubnetsStr := terraform.OutputList(t, terraformOptions, "public_subnets")

			// Verify number of subnets
			assert.Equal(t, 3, len(privateSubnetsStr), "Should have 3 private subnets")
			assert.Equal(t, 3, len(publicSubnetsStr), "Should have 3 public subnets")

			// Convert subnet IDs to AWS SDK format
			privateSubnetIDs := make([]*string, len(privateSubnetsStr))
			publicSubnetIDs := make([]*string, len(publicSubnetsStr))
			for i, id := range privateSubnetsStr {
				privateSubnetIDs[i] = aws.String(id)
			}
			for i, id := range publicSubnetsStr {
				publicSubnetIDs[i] = aws.String(id)
			}

			// Check private subnets
			privateSubnets, err := ec2Client.DescribeSubnets(&ec2.DescribeSubnetsInput{
				SubnetIds: privateSubnetIDs,
			})
			require.NoError(t, err)

			privateAZs := make(map[string]bool)
			for _, subnet := range privateSubnets.Subnets {
				// Print all tags for debugging
				fmt.Printf("Private subnet tags:\n")
				for _, tag := range subnet.Tags {
					fmt.Printf("  %s: %s\n", *tag.Key, *tag.Value)
				}

				// Verify subnet tags
				expectedTags := map[string]string{
					"Environment": tc.environment,
					"Project":     projectName,
					"ManagedBy":   "terraform",
				}
				if !hasRequiredTags(subnet.Tags, expectedTags) {
					fmt.Printf("Expected tags: %v\n", expectedTags)
				}
				assert.True(t, hasRequiredTags(subnet.Tags, expectedTags),
					"Private subnet should have all required tags")

				// Verify subnet is private (no auto-assign public IP)
				assert.False(t, *subnet.MapPublicIpOnLaunch)
				privateAZs[*subnet.AvailabilityZone] = true
			}

			// Check public subnets
			publicSubnets, err := ec2Client.DescribeSubnets(&ec2.DescribeSubnetsInput{
				SubnetIds: publicSubnetIDs,
			})
			require.NoError(t, err)

			publicAZs := make(map[string]bool)
			for _, subnet := range publicSubnets.Subnets {
				fmt.Printf("Public subnet tags:\n")
				for _, tag := range subnet.Tags {
					fmt.Printf("  %s: %s\n", *tag.Key, *tag.Value)
				}

				expectedTags := map[string]string{
					"Environment": tc.environment,
					"Project":     projectName,
					"ManagedBy":   "terraform",
				}
				if !hasRequiredTags(subnet.Tags, expectedTags) {
					fmt.Printf("Expected tags: %v\n", expectedTags)
				}
				assert.True(t, hasRequiredTags(subnet.Tags, expectedTags),
					"Public subnet should have all required tags")

				// Verify subnet is public (auto-assign public IP)
				assert.True(t, *subnet.MapPublicIpOnLaunch)
				publicAZs[*subnet.AvailabilityZone] = true
			}

			// Verify subnets are in different AZs
			assert.Equal(t, 3, len(privateAZs), "Private subnets should be in different AZs")
			assert.Equal(t, 3, len(publicAZs), "Public subnets should be in different AZs")

			// Check NAT Gateways
			natGateways, err := ec2Client.DescribeNatGateways(&ec2.DescribeNatGatewaysInput{
				Filter: []*ec2.Filter{
					{
						Name:   aws.String("vpc-id"),
						Values: []*string{aws.String(vpcID)},
					},
				},
			})
			require.NoError(t, err)

			if tc.environment == "prod" {
				assert.Equal(t, 3, len(natGateways.NatGateways), "Production should have one NAT Gateway per AZ")
			} else {
				assert.Equal(t, 1, len(natGateways.NatGateways), "Non-production should have a single NAT Gateway")
			}

			// Verify VPC attributes
			describeVpcAttributeInput := &ec2.DescribeVpcAttributeInput{
				VpcId: aws.String(vpcID),
			}

			// Check DNS hostnames
			describeVpcAttributeInput.Attribute = aws.String("enableDnsHostnames")
			dnsHostnames, err := ec2Client.DescribeVpcAttribute(describeVpcAttributeInput)
			require.NoError(t, err)
			assert.True(t, *dnsHostnames.EnableDnsHostnames.Value, "VPC should have DNS hostnames enabled")

			// Check DNS support
			describeVpcAttributeInput.Attribute = aws.String("enableDnsSupport")
			dnsSupport, err := ec2Client.DescribeVpcAttribute(describeVpcAttributeInput)
			require.NoError(t, err)
			assert.True(t, *dnsSupport.EnableDnsSupport.Value, "VPC should have DNS support enabled")

			// Check flow logs
			flowLogs, err := ec2Client.DescribeFlowLogs(&ec2.DescribeFlowLogsInput{
				Filter: []*ec2.Filter{
					{
						Name:   aws.String("resource-id"),
						Values: []*string{aws.String(vpcID)},
					},
				},
			})
			require.NoError(t, err)
			assert.NotEmpty(t, flowLogs.FlowLogs, "VPC should have flow logs enabled")
		})
	}
}
