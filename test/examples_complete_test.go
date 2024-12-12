package test

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"context"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/wait"
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

func TestTerraformExampleComplete(t *testing.T) {
	// Configure LocalStack
	ctx := context.Background()

	// Start LocalStack Container
	localstackAuthToken := getLocalStackAuthToken(t)
	localstackContainer := startLocalStack(ctx, t, localstackAuthToken)
	defer terminateContainer(ctx, localstackContainer)

	// Retrieve LocalStack Endpoint
	endpoint := getContainerEndpoint(ctx, t, localstackContainer)
	t.Logf("LocalStack endpoint: %s", endpoint)

	// Setup Temporary AWS Profile
	configPath, credentialsPath := setupAWSProfile(t)
	setAWSEnvVars(configPath, credentialsPath, endpoint)

	//ec2Client := createAWSClient(t, endpoint, credentialsPath)
	//t.Logf("AWS EC2 client created successfully")

	// Define the original module and create a temp test folder
	originalModuleDir := filepath.Join("..")

	tempTestFolder := test_structure.CopyTerraformFolderToTemp(t, originalModuleDir, path.Join("examples", "complete"))

	generateProviderConfig(t, tempTestFolder, endpoint)

	// Save the Terraform options for reuse across stages
	terraformOptions := &terraform.Options{
		TerraformDir: tempTestFolder,
		NoColor:      true,
	}

	test_structure.SaveTerraformOptions(t, tempTestFolder, terraformOptions)

	// Stage 1: Init and Apply
	defer test_structure.RunTestStage(t, "cleanup", func() {
		t.Log("Running terraform destroy...")
		terraform.Destroy(t, terraformOptions)
	})

	test_structure.RunTestStage(t, "setup", func() {
		t.Log("Running terraform init and apply...")
		terraform.InitAndApply(t, terraformOptions)
	})

	// Stage 2: Validate Outputs
	test_structure.RunTestStage(t, "validate", func() {
		t.Log("Validating Terraform outputs and SSM parameters...")

		// Verify Terraform output
		output := terraform.Output(t, terraformOptions, "ssm_parameter_paths")
		assert.NotEmpty(t, output, "Terraform output 'ssm_parameter_paths' should not be empty")

		// Initialize AWS SSM client
		ssmClient := createSSMClient(t, endpoint, credentialsPath)

		// List of expected SSM parameter paths
		expectedParameters := map[string]string{
			"/dev/krabby/API_KEY":        "api-XXXXXXXXXXXXXXXXXXXXX",
			"/dev/krabby/S3_BUCKET_ARN":  "",
			"/dev/krabby/S3_BUCKET_NAME": "",
		}

		for param, expectedValue := range expectedParameters {
			t.Logf("Verifying SSM parameter: %s", param)
			actualValue := getSSMParameter(t, ssmClient, param)

			// Validate the value for API_KEY, others just check existence
			if param == "/dev/krabby/API_KEY" {
				assert.Equal(t, expectedValue, actualValue, "SSM parameter value mismatch for API_KEY")
			} else {
				assert.NotEmpty(t, actualValue, fmt.Sprintf("SSM parameter %s should not be empty", param))
			}
		}
	})
}

// ---------------- Utility Functions ----------------

// setupAWSProfile creates a temporary AWS profile
func setupAWSProfile(t *testing.T) (string, string) {
	tmpDir, _ := os.MkdirTemp("", "awsconfig")
	credentialsPath := filepath.Join(tmpDir, "credentials")
	configPath := filepath.Join(tmpDir, "config")

	// Write credentials
	_ = os.WriteFile(credentialsPath, []byte(`[localstack]
aws_access_key_id = test
aws_secret_access_key = test`), 0644)
	t.Logf("AWS credentials created successfully in %s", credentialsPath)
	// Write config
	_ = os.WriteFile(configPath, []byte(`[profile localstack]
region = us-east-1
output = json`), 0644)
	t.Logf("AWS profile created successfully in %s", configPath)
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
func getLocalStackAuthToken(t *testing.T) string {
	token := os.Getenv("LOCALSTACK_AUTH_TOKEN")
	// Don't fail if token is not set for non-pro version
	//if token == "" {
	//	t.Fatalf("LOCALSTACK_AUTH_TOKEN is not set")
	//}
	return token
}

// startLocalStack starts the LocalStack container with appropriate port bindings and configurations.
func startLocalStack(ctx context.Context, t *testing.T, authToken string) testcontainers.Container {
	// Define LocalStack ports to expose
	ports := []string{
		"4566/tcp", // LocalStack main port
		"443/tcp",  // HTTPS port
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
		t.Fatalf("Failed to start LocalStack cnt: %v", err)
	}

	return cnt
}

// terminateContainer ensures the container is terminated
func terminateContainer(ctx context.Context, container testcontainers.Container) {
	_ = container.Terminate(ctx)
}

// getContainerEndpoint retrieves LocalStack endpoint
func getContainerEndpoint(ctx context.Context, t *testing.T, container testcontainers.Container) string {
	endpoint, err := container.PortEndpoint(ctx, localstackPort, "http")
	if err != nil {
		t.Fatalf("Failed to get endpoint: %v", err)
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
    cloudformation = "http://localhost:4566"      # CloudFormation endpoint
	ssm            = "http://localhost:4566"       # SSM (Parameter Store) endpoint
  }
}`)

	providerConfigPath := path.Join(dir, "provider.tf")
	err := os.WriteFile(path.Join(providerConfigPath), []byte(providerConfig), 0644)
	if err != nil {
		t.Errorf("Failed to write provider config: %v", err)
	}

	t.Logf("Provider LocalStack configured in %s", providerConfigPath)

	// Outputs
	outputsConfig := fmt.Sprintf(`
	output "ssm_parameter_paths" {
	  value = module.krabby.ssm_parameter_paths[*]
	}
	`)

	outputsConfigPath := path.Join(dir, "outputs.tf")
	err = os.WriteFile(path.Join(outputsConfigPath), []byte(outputsConfig), 0644)
	if err != nil {
		t.Errorf("Failed to write provider config: %v", err)
	}

	t.Logf("Provider LocalStack configured in %s", outputsConfigPath)
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
