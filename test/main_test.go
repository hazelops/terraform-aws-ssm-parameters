/*
 * SPDX-License-Identifier: Apache-2.0
 * SPDX-FileCopyrightText: © 2024 HazelOps OÜ
 */

package test

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

// Constants for LocalStack and AWS
const (
	localstackImage = "localstack/localstack:4.0.3"
	//localstackImage    = "localstack/localstack-pro:4.0.3"
	localstackPort     = "4566/tcp"
	testAwsProfile     = "localstack"
	testAwsRegion      = "us-east-1"
	localstackReadyLog = "Ready."
)

var (
	localstackEndpoint, configPath, credentialsPath string
)

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	// Configure LocalStack
	ctx := context.Background()

	// Start LocalStack Container
	localstackAuthToken := getLocalStackAuthToken()
	localstackContainer := startLocalStack(ctx, localstackAuthToken)
	defer terminateContainer(ctx, localstackContainer)

	// Retrieve LocalStack Endpoint
	localstackEndpoint = getContainerEndpoint(ctx, localstackContainer)
	log.Printf("LocalStack endpoint: %s", localstackEndpoint)

	// Setup Temporary AWS Profile
	configPath, credentialsPath = setupAWSProfile()
	setAWSEnvVars(configPath, credentialsPath, localstackEndpoint)
	os.Exit(m.Run())
}

// ---------------- Utility Functions ----------------

// setupAWSProfile creates a temporary AWS profile
func setupAWSProfile() (string, string) {
	tmpDir, _ := os.MkdirTemp("", "awsconfig")
	credentialsPath = filepath.Join(tmpDir, "credentials")
	configPath = filepath.Join(tmpDir, "config")

	// Write credentials
	_ = os.WriteFile(credentialsPath, []byte(`[localstack]
aws_access_key_id = test
aws_secret_access_key = test`), 0644)
	log.Printf("AWS credentials created successfully in %s", credentialsPath)
	// Write config
	_ = os.WriteFile(configPath, []byte(`[profile localstack]
region = us-east-1
output = json`), 0644)
	log.Printf("AWS profile created successfully in %s", configPath)
	return configPath, credentialsPath
}

// setAWSEnvVars sets AWS environment variables
func setAWSEnvVars(configPath, credentialsPath, endpoint string) {
	os.Setenv("AWS_PROFILE", testAwsProfile)
	os.Setenv("AWS_CONFIG_FILE", configPath)
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", credentialsPath)
	os.Setenv("AWS_REGION", testAwsRegion)
	os.Setenv("AWS_ENDPOINT_URL", endpoint)
}

// getLocalStackAuthToken retrieves LocalStack API token
func getLocalStackAuthToken() string {
	token := os.Getenv("LOCALSTACK_AUTH_TOKEN")
	// Don't fail if token is not set for non-pro version
	//if token == "" {
	//	t.Fatalf("LOCALSTACK_AUTH_TOKEN is not set")
	//}
	return token
}

// startLocalStack starts the LocalStack container with appropriate port bindings and configurations.
func startLocalStack(ctx context.Context, authToken string) testcontainers.Container {
	// Define LocalStack ports to expose
	ports := []string{
		"4566/tcp", // LocalStack main port
		//"443/tcp",  // HTTPS port
	}

	// Add individual ports for the range 4510-4559
	for i := 4510; i <= 4559; i++ {
		ports = append(ports, fmt.Sprintf("%d/tcp", i))
	}

	// Define port bindings dynamically
	portBindings := make(map[nat.Port][]nat.PortBinding)
	for _, p := range ports {
		port := nat.Port(p)
		portBindings[port] = []nat.PortBinding{
			{HostIP: "127.0.0.1", HostPort: port.Port()},
		}
	}

	// Define cnt request
	req := testcontainers.ContainerRequest{
		Image:        localstackImage,
		ExposedPorts: ports,
		Env: map[string]string{
			"LOCALSTACK_AUTH_TOKEN": authToken,
		},
		Mounts: testcontainers.Mounts(
			testcontainers.BindMount("/var/run/docker.sock", "/var/run/docker.sock"),
		),
		WaitingFor: wait.ForLog(localstackReadyLog),
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.PortBindings = portBindings

		},
	}

	// Start the container
	cnt, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Failed to start LocalStack cnt: %v", err)
	}

	return cnt
}

// terminateContainer ensures the container is terminated
func terminateContainer(ctx context.Context, container testcontainers.Container) {
	_ = container.Terminate(ctx)
}

// getContainerEndpoint retrieves LocalStack endpoint
func getContainerEndpoint(ctx context.Context, container testcontainers.Container) string {
	endpoint, err := container.PortEndpoint(ctx, localstackPort, "http")
	if err != nil {
		log.Fatalf("Failed to get endpoint: %v", err)
	}
	return endpoint
}

// createAWSClient initializes AWS EC2 client
func createAWSClient(t *testing.T, endpoint, credentialsPath string) *ec2.EC2 {
	cfg := &aws.Config{
		Region:      aws.String(testAwsRegion),
		Endpoint:    aws.String(endpoint),
		Credentials: credentials.NewSharedCredentials(credentialsPath, testAwsProfile),
		DisableSSL:  aws.Bool(true),
	}
	sess, _ := session.NewSession(cfg)
	return ec2.New(sess)
}

func generateProviderConfig(t *testing.T, dir, endpoint string) {
	// Generate Terraform configuration
	providerConfig := fmt.Sprintf(`provider "aws" {
  region                      = "us-east-1"    # Use any region
  access_key                  = "test"         # Dummy credentials for LocalStack
  secret_key                  = "test"
  s3_use_path_style = true # Required for LocalStack
  skip_credentials_validation = true           # Skip AWS credential checks
  skip_requesting_account_id  = true           # Skip AWS account check

  endpoints {
    s3             = "http://localhost:4566"       # S3 endpoint
    dynamodb       = "http://localhost:4566"       # DynamoDB endpoint
    sqs            = "http://localhost:4566"       # SQS endpoint
    sns            = "http://localhost:4566"       # SNS endpoint
    ec2            = "http://localhost:4566"       # EC2 endpoint
    cloudwatch     = "http://localhost:4566"       # CloudWatch endpoint
    sts            = "http://localhost:4566"       # STS endpoint
    iam            = "http://localhost:4566"       # IAM endpoint
    lambda         = "http://localhost:4566"       # Lambda endpoint
    cloudformation = "http://localhost:4566"       # CloudFormation endpoint
	ssm            = "http://localhost:4566"       # SSM (Parameter Store) endpoint
  }
}`)

	providerConfigPath := path.Join(dir, "provider.tf")
	err := os.WriteFile(path.Join(providerConfigPath), []byte(providerConfig), 0644)
	if err != nil {
		t.Errorf("Failed to write provider config: %v", err)
	}

	t.Logf("Provider LocalStack configured in %s", providerConfigPath)
}

func createSSMClient(t *testing.T, endpoint, credentialsPath string) *ssm.SSM {
	cfg := &aws.Config{
		Region:      aws.String(testAwsRegion),
		Endpoint:    aws.String(endpoint),
		Credentials: credentials.NewSharedCredentials(credentialsPath, testAwsProfile),
		DisableSSL:  aws.Bool(true),
	}
	sess, err := session.NewSession(cfg)
	if err != nil {
		t.Fatalf("Failed to create AWS session: %v", err)
	}
	return ssm.New(sess)
}

func getSSMParameter(t *testing.T, ssmClient *ssm.SSM, parameterName string) string {
	param, err := ssmClient.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(parameterName),
		WithDecryption: aws.Bool(false),
	})
	if err != nil {
		t.Fatalf("Failed to retrieve SSM parameter %s: %v", parameterName, err)
	}
	value := *param.Parameter.Value

	// Strip LocalStack's prefix if it exists
	if strings.HasPrefix(value, "kms:alias/aws/ssm:") {
		value = strings.TrimPrefix(value, "kms:alias/aws/ssm:")
	}
	return value
}
