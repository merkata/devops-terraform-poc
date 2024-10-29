// test/utils/aws.go
package utils

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/gruntwork-io/terratest/modules/terraform"
)

// CreateSession creates an AWS session for the specified region
func CreateSession(region string) *session.Session {
	return session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
}

// CreateEC2Client creates an EC2 client
func CreateEC2Client(region string) *ec2.EC2 {
	return ec2.New(CreateSession(region))
}

// CreateASGClient creates an AutoScaling client
func CreateASGClient(region string) *autoscaling.AutoScaling {
	return autoscaling.New(CreateSession(region))
}

// CreateIAMClient creates an IAM client
func CreateIAMClient(region string) *iam.IAM {
	return iam.New(CreateSession(region))
}

// Tag represents a key-value pair tag
type Tag struct {
	Key   string
	Value string
}

// EC2Tag represents an EC2 tag
type EC2Tag struct {
	Tag *ec2.Tag
}

// ConvertEC2TagsToTags converts EC2 tags to our Tag structure
func ConvertEC2TagsToTags(ec2Tags []*ec2.Tag) []Tag {
	var tags []Tag
	for _, tag := range ec2Tags {
		tags = append(tags, Tag{
			Key:   aws.StringValue(tag.Key),
			Value: aws.StringValue(tag.Value),
		})
	}
	return tags
}

// HasRequiredTags checks if all required tags are present in the given tags
func HasRequiredTags(actual []Tag, required map[string]string) bool {
	foundTags := make(map[string]bool)
	for _, tag := range actual {
		if expectedValue, ok := required[tag.Key]; ok {
			if expectedValue == tag.Value {
				foundTags[tag.Key] = true
			}
		}
	}

	// Check if all required tags were found with correct values
	for key := range required {
		if !foundTags[key] {
			return false
		}
	}

	return true
}

// CreateALB creates a test ALB configuration
func CreateALB(t TestingT, region, environment, projectName string, vpcID string, publicSubnets []string) *terraform.Options {
	return &terraform.Options{
		TerraformDir: "../modules/alb",
		Vars: map[string]interface{}{
			"environment":     environment,
			"project_name":    projectName,
			"vpc_id":          vpcID,
			"public_subnets":  publicSubnets,
			"certificate_arn": "arn:aws:acm:us-east-1:683721267198:certificate/aa67a8ae-f2fe-4cef-95e6-a676fd11f5be", // Replace with a valid certificate ARN for testing
			"apps": map[string]interface{}{
				"app1": map[string]interface{}{
					"port":             8085,
					"path":             "/app1/*",
					"health_check_url": "/app1/status",
					"domain":           []string{"merkata.cloudns.be"},
					"priority":         100,
				},
				"app2": map[string]interface{}{
					"port":             8086,
					"path":             "/app2/*",
					"health_check_url": "/app2/status",
					"domain":           []string{"merkata.cloudns.be"},
					"priority":         200,
				},
			},
		},
		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": region,
		},
	}
}

// CreateVPCTestConfig creates a test VPC configuration
func CreateVPC(t TestingT, region, environment, projectName string) *terraform.Options {
	return &terraform.Options{
		TerraformDir: "../modules/vpc",
		Vars: map[string]interface{}{
			"environment":  environment,
			"project_name": projectName,
			"vpc_cidr":     "10.0.0.0/16",
		},
		EnvVars: map[string]string{
			"AWS_DEFAULT_REGION": region,
		},
	}
}

// TestingT is an interface wrapper around testing.T
type TestingT interface {
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}
