package test

import (
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/azure"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

// You normally want to run this under a separate "Testing" subscription
// For lab purposes you will use your assigned subscription under the Cloud Dev/Ops program tenant
var subscriptionID string = "c33d7468-11ea-4d59-906e-b5ab730587f8"

func TestAzureLinuxVMCreation(t *testing.T) {
	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: "../",
		// Override the default terraform variables
		Vars: map[string]interface{}{
			"labelPrefix": "seto0046",
		},
		MaxRetries:         3,
		TimeBetweenRetries: 5 * time.Second,
		RetryableTerraformErrors: map[string]string{
			"NetworkSecurityGroupOldReferencesNotCleanedUp": "Waiting for NSG references to be cleaned up",
			"InternalServerError":                           "Waiting for Azure internal server error to resolve",
		},
	}

	defer terraform.Destroy(t, terraformOptions)

	// Run `terraform init` and `terraform apply`. Fail the test if there are any errors.
	terraform.InitAndApply(t, terraformOptions)

	// Run `terraform output` to get the value of output variable
	vmName := terraform.Output(t, terraformOptions, "vm_name")
	resourceGroupName := terraform.Output(t, terraformOptions, "resource_group_name")

	// Confirm VM exists
	assert.True(t, azure.VirtualMachineExists(t, vmName, resourceGroupName, subscriptionID))

	// Test 1: Confirm NIC exists and is connected to the VM
	nics := azure.GetVirtualMachineNics(t, vmName, resourceGroupName, subscriptionID)
	assert.NotEmpty(t, nics, "VM should have at least one NIC attached")

	// Test 2: Confirm VM is running the correct Ubuntu version
	vmImage := azure.GetVirtualMachineImage(t, vmName, resourceGroupName, subscriptionID)
	assert.Equal(t, "Canonical", vmImage.Publisher, "VM should be running Ubuntu")
	assert.Equal(t, "0001-com-ubuntu-server-jammy", vmImage.Offer, "VM should be running Ubuntu Server")
	assert.Equal(t, "22_04-lts-gen2", vmImage.SKU, "VM should be running Ubuntu 18.04 LTS")
}
