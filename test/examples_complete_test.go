/*
 * SPDX-License-Identifier: Apache-2.0
 * SPDX-FileCopyrightText: © 2024 HazelOps OÜ
 */

package test

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestTerraformExampleComplete(t *testing.T) {
	//t.Parallel()
	exampleName := "complete"

	// Define the original module and create a temp test folder
	originalModuleDir := filepath.Join("..")

	tempTestFolder := test_structure.CopyTerraformFolderToTemp(t, originalModuleDir, path.Join("examples", exampleName))

	generateProviderConfig(t, tempTestFolder, localstackEndpoint)

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

		writeOutputs(t, tempTestFolder)

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
		ssmClient := createSSMClient(t, localstackEndpoint, credentialsPath)

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

func writeOutputs(t *testing.T, dir string) {
	// Outputs
	outputsConfig := fmt.Sprintf(`
	output "ssm_parameter_paths" {
	  value = module.krabby.ssm_parameter_paths[*]
	}
	`)

	outputsConfigPath := path.Join(dir, "outputs.tf")
	err := os.WriteFile(path.Join(outputsConfigPath), []byte(outputsConfig), 0644)
	if err != nil {
		t.Errorf("Failed to write outputs: %v", err)
	}

	t.Logf("Outputs configured in %s", outputsConfigPath)
}
