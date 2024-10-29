package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestALBModule(t *testing.T) {
	t.Parallel()

	// Use the same workingDir for both stages
	workingDir := test_structure.CopyTerraformFolderToTemp(t, "../", "modules/alb")

	// Generate a random name to prevent a naming conflict
	uniqueID := strings.ToLower(random.UniqueId()) // The module lowercases the project name
	projectName := fmt.Sprintf("%s", uniqueID)

	// Test cases
	testCases := []struct {
		name           string
		region         string
		environment    string
		certificateArn string
		apps           map[string]interface{}
	}{
		{
			name:           "us-east-1-ci",
			region:         "us-east-1",
			environment:    "ci",
			certificateArn: "arn:aws:acm:us-east-1:683721267198:certificate/aa67a8ae-f2fe-4cef-95e6-a676fd11f5be", // Replace with a valid certificate ARN for testing
			apps: map[string]interface{}{
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
	}

	for _, testCase := range testCases {
		tc := testCase // Required to avoid variable capture in closure

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create VPC first as ALB depends on it
			vpcOpts := createVPC(t, tc.region, tc.environment, projectName)
			defer terraform.Destroy(t, vpcOpts)

			// Deploy VPC and get outputs
			terraform.InitAndApply(t, vpcOpts)
			vpcID := terraform.Output(t, vpcOpts, "vpc_id")
			publicSubnets := terraform.OutputList(t, vpcOpts, "public_subnets")

			// Set up ALB options
			terraformOptions := &terraform.Options{
				TerraformDir: workingDir,
				Vars: map[string]interface{}{
					"environment":     tc.environment,
					"project_name":    projectName,
					"vpc_id":          vpcID,
					"public_subnets":  publicSubnets,
					"certificate_arn": tc.certificateArn,
					"apps":            tc.apps,
				},
				EnvVars: map[string]string{
					"AWS_DEFAULT_REGION": tc.region,
				},
			}

			// Clean up resources when the test finishes
			defer terraform.Destroy(t, terraformOptions)

			// Deploy the ALB
			terraform.InitAndApply(t, terraformOptions)

			// Get outputs
			albDNSName := terraform.Output(t, terraformOptions, "alb_dns_name")
			albName := terraform.Output(t, terraformOptions, "alb_name")
			targetGroupArns := terraform.OutputMap(t, terraformOptions, "target_group_arns")
			albSGID := terraform.Output(t, terraformOptions, "alb_security_group_id")

			// Create AWS ELBv2 client
			elbv2Client := createELBv2Client(tc.region)

			// Test ALB Configuration
			testALBConfiguration(t, elbv2Client, albDNSName, tc.environment, projectName)

			// Test Target Groups
			testTargetGroups(t, elbv2Client, targetGroupArns, tc.apps, vpcID)

			// Test Listener Rules
			testListenerRules(t, elbv2Client, albName, tc.apps)

			// Test Security Group Rules
			testSecurityGroupRules(t, tc.region, albSGID, tc.apps)
		})
	}
}

func createVPC(t *testing.T, region, environment, projectName string) *terraform.Options {
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

func createELBv2Client(region string) *elbv2.ELBV2 {
	return elbv2.New(createSession(region))
}

func testALBConfiguration(t *testing.T, client *elbv2.ELBV2, albDNSName, environment, projectName string) {
	// Get ALB by DNS name
	input := &elbv2.DescribeLoadBalancersInput{
		Names: []*string{aws.String(fmt.Sprintf("%s-%s-alb", projectName, environment))},
	}

	result, err := client.DescribeLoadBalancers(input)
	require.NoError(t, err)
	require.Len(t, result.LoadBalancers, 1)

	alb := result.LoadBalancers[0]

	// Verify ALB configuration
	assert.Equal(t, "application", *alb.Type)
	assert.False(t, *alb.Scheme == "internal")
	assert.Equal(t, albDNSName, *alb.DNSName)

	// Check ALB Tags
	tagsInput := &elbv2.DescribeTagsInput{
		ResourceArns: []*string{alb.LoadBalancerArn},
	}

	tagsOutput, err := client.DescribeTags(tagsInput)
	require.NoError(t, err)
	require.Len(t, tagsOutput.TagDescriptions, 1)

	hasExpectedTags := hasRequiredELBTags(tagsOutput.TagDescriptions[0].Tags, map[string]string{
		"Environment": environment,
		"Project":     projectName,
		"ManagedBy":   "terraform",
	})
	assert.True(t, hasExpectedTags)
}

func testTargetGroups(t *testing.T, client *elbv2.ELBV2, targetGroupArns map[string]string, apps map[string]interface{}, vpcID string) {
	for appName, arn := range targetGroupArns {
		input := &elbv2.DescribeTargetGroupsInput{
			TargetGroupArns: []*string{aws.String(arn)},
		}

		result, err := client.DescribeTargetGroups(input)
		require.NoError(t, err)
		require.Len(t, result.TargetGroups, 1)

		tg := result.TargetGroups[0]
		app := apps[appName].(map[string]interface{})

		// Verify target group configuration
		assert.Equal(t, vpcID, *tg.VpcId)
		assert.Equal(t, "HTTP", *tg.Protocol)
		assert.Equal(t, int64(app["port"].(int)), *tg.Port)

		// Verify health check configuration
		assert.Equal(t, app["health_check_url"].(string), *tg.HealthCheckPath)
		assert.Equal(t, int64(3), *tg.HealthyThresholdCount)
		assert.Equal(t, int64(3), *tg.UnhealthyThresholdCount)
	}
}

func testListenerRules(t *testing.T, client *elbv2.ELBV2, albName string, apps map[string]interface{}) {
	// Get ALB Listeners
	input := &elbv2.DescribeLoadBalancersInput{
		Names: []*string{aws.String(albName)},
	}

	result, err := client.DescribeLoadBalancers(input)
	require.NoError(t, err)
	require.Len(t, result.LoadBalancers, 1)

	listenersInput := &elbv2.DescribeListenersInput{
		LoadBalancerArn: result.LoadBalancers[0].LoadBalancerArn,
	}

	listeners, err := client.DescribeListeners(listenersInput)
	require.NoError(t, err)

	// Find HTTPS listener
	var httpsListener *elbv2.Listener
	for _, listener := range listeners.Listeners {
		if *listener.Protocol == "HTTPS" {
			httpsListener = listener
			break
		}
	}
	require.NotNil(t, httpsListener)

	// Check listener rules
	rulesInput := &elbv2.DescribeRulesInput{
		ListenerArn: httpsListener.ListenerArn,
	}

	rules, err := client.DescribeRules(rulesInput)
	require.NoError(t, err)

	// Verify rules match app configuration
	for appName, appConfig := range apps {
		app := appConfig.(map[string]interface{})
		found := false

		for _, rule := range rules.Rules {
			if rule == nil {
				continue
			}
			if rule.Priority == nil {
				continue
			}
			found = true

			// Helper function to safely check conditions
			findCondition := func(conditionType string) *elbv2.RuleCondition {
				for _, condition := range rule.Conditions {
					if condition != nil && condition.Field != nil && *condition.Field == conditionType {
						return condition
					}
				}
				return nil
			}

			// Check path pattern
			pathCondition := findCondition("path-pattern")
			require.NotNil(t, pathCondition, "Path pattern condition not found for app %s", appName)
			require.NotNil(t, pathCondition.PathPatternConfig, "Path pattern config is nil for app %s", appName)
			require.NotEmpty(t, pathCondition.PathPatternConfig.Values, "Path pattern values are empty for app %s", appName)
			assert.Equal(t, app["path"].(string), *pathCondition.PathPatternConfig.Values[0],
				"Path pattern mismatch for app %s", appName)

			// Check host header
			hostCondition := findCondition("host-header")
			require.NotNil(t, hostCondition, "Host header condition not found for app %s", appName)
			require.NotNil(t, hostCondition.HostHeaderConfig, "Host header config is nil for app %s", appName)
			require.NotEmpty(t, hostCondition.HostHeaderConfig.Values, "Host header values are empty for app %s", appName)
			assert.Equal(t, app["domain"].([]string)[0], *hostCondition.HostHeaderConfig.Values[0],
				"Host header mismatch for app %s", appName)

			break
		}
		assert.True(t, found, "No rule found for app %s", appName)
	}
}

func testSecurityGroupRules(t *testing.T, region, sgID string, apps map[string]interface{}) {
	ec2Client := createEC2Client(region)

	input := &ec2.DescribeSecurityGroupsInput{
		GroupIds: []*string{aws.String(sgID)},
	}

	result, err := ec2Client.DescribeSecurityGroups(input)
	require.NoError(t, err)
	require.Len(t, result.SecurityGroups, 1)

	sg := result.SecurityGroups[0]

	// Verify inbound rules
	foundHTTP := false
	foundHTTPS := false
	for _, rule := range sg.IpPermissions {
		if *rule.FromPort == 80 {
			foundHTTP = true
			assert.Equal(t, int64(80), *rule.ToPort)
			assert.Equal(t, "tcp", *rule.IpProtocol)
		}
		if *rule.FromPort == 443 {
			foundHTTPS = true
			assert.Equal(t, int64(443), *rule.ToPort)
			assert.Equal(t, "tcp", *rule.IpProtocol)
		}
	}

	assert.True(t, foundHTTP, "HTTP rule not found")
	assert.True(t, foundHTTPS, "HTTPS rule not found")

	// Verify outbound rules
	require.Len(t, sg.IpPermissionsEgress, 1)
	assert.Equal(t, "-1", *sg.IpPermissionsEgress[0].IpProtocol) // All traffic
}

func hasRequiredELBTags(tags []*elbv2.Tag, requiredTags map[string]string) bool {
	tagMap := make(map[string]string)
	for _, tag := range tags {
		tagMap[*tag.Key] = *tag.Value
	}

	for key, value := range requiredTags {
		if tagMap[key] != value {
			return false
		}
	}
	return true
}

func createSession(region string) *session.Session {
	return session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
}

func createEC2Client(region string) *ec2.EC2 {
	return ec2.New(createSession(region))
}
