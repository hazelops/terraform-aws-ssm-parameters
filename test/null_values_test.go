/*
 * SPDX-License-Identifier: Apache-2.0
 * SPDX-FileCopyrightText: © 2024 HazelOps OÜ
 */

package test

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
)

func TestTerraformNullValues(t *testing.T) {
	//t.Parallel()
	exampleName := "null-values"

	//ec2Client := createAWSClient(t, endpoint, credentialsPath)
	//t.Logf("AWS EC2 client created successfully")

	// Define the original module and create a temp test folder
	originalModuleDir := filepath.Join("..")

	generateTestModuleCallNullValues(t, originalModuleDir)

	tempTestFolder := test_structure.CopyTerraformFolderToTemp(t, originalModuleDir, path.Join("examples", exampleName))

	generateProviderConfig(t, tempTestFolder, localstackEndpoint)

	// Save the Terraform options for reuse across stages
	terraformOptions := &terraform.Options{
		TerraformDir: tempTestFolder,
		NoColor:      true,
	}

	test_structure.SaveTerraformOptions(t, tempTestFolder, terraformOptions)

	// Stage 1: Setup
	defer test_structure.RunTestStage(t, "cleanup", func() {
		t.Log("Running terraform destroy...")
		out, err := terraform.DestroyE(t, terraformOptions)
		if err != nil {
			t.Logf("Failed to run terraform destroy (and it's expected in this test): %v", err)
		}
		t.Logf("Terraform destroy stdout: %s", out)
	})

	test_structure.RunTestStage(t, "setup", func() {
		t.Log("Running terraform init and apply...")

		out, err := terraform.InitAndApplyE(t, terraformOptions)
		if err != nil {
			t.Logf("Failed to run terraform init and apply (and it's expected in this test): %v", err)
		}

		// Log the captured stdout
		t.Logf("Terraform stdout: %s", out)
		// Save the output for validation stage
		test_structure.SaveString(t, tempTestFolder, "terraformOutput", out)
	})

	// Stage 2: Validate Outputs
	test_structure.RunTestStage(t, "validate", func() {
		t.Log("Validating Terraform output...")

		// Retrieve the saved output
		out := test_structure.LoadString(t, tempTestFolder, "terraformOutput")

		assert.Contains(t, out, "Error: Invalid value for variable", "Terraform output should contain 'Error: Invalid value for variable'")
		assert.Contains(t, out, "All values in the 'parameters' map must be non-null", "Terraform output should contain 'All values in the 'parameters' map must be non-null'")
	})
}

func generateTestModuleCallNullValues(t *testing.T, destFolder string) {
	// Create a module directory
	testModuleFolder := filepath.Join(destFolder, "examples", "null-values")
	err := os.MkdirAll(testModuleFolder, 0755)
	if err != nil {
		t.Fatalf("Error creating directory: %s", err)
	}

	// Generate a test module call with null values
	moduleCall := `
module "null_values" {
  source = "../../"
  env    = "dev"
  name   = "null-values"

  parameters = {
    API_KEY        = null
    S3_BUCKET_ARN  = "aws:s3:::null-values"
    S3_BUCKET_NAME = "null-values"
  }
}`

	// Write the module call to the main.tf file
	err = os.WriteFile(filepath.Join(testModuleFolder, "test.tf"), []byte(moduleCall), 0644)
	if err != nil {
		t.Fatalf("Error writing to main.tf: %s", err)
	}
}
