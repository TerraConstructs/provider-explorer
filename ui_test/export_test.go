package ui_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	tfjson "github.com/hashicorp/terraform-json"

	"github.com/terraconstructs/provider-explorer/internal/ui"
)

func Test_HCL_Export_Arguments_To_Variables(t *testing.T) {
	// Set consistent color profile for stable output
	lipgloss.SetColorProfile(0)

	// Load minimal AWS fixture
	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	// Get aws_instance schema directly
	providerSchema := ps.Schemas["registry.terraform.io/hashicorp/aws"]
	instanceSchema := providerSchema.ResourceSchemas["aws_instance"]

	// Test arguments to variables conversion
	result := ui.ConvertArgumentsToHCLVariables(instanceSchema)

	// Verify result contains expected variable blocks
	if !strings.Contains(result, `variable "ami"`) {
		t.Errorf("Expected ami variable, got: %s", result)
	}

	if !strings.Contains(result, `variable "instance_type"`) {
		t.Errorf("Expected instance_type variable, got: %s", result)
	}

	if !strings.Contains(result, `type = any`) {
		t.Errorf("Expected type declarations, got: %s", result)
	}

	if !strings.Contains(result, `description = "Required argument`) {
		t.Errorf("Expected required argument descriptions, got: %s", result)
	}

	// Should NOT contain computed attributes
	if strings.Contains(result, `variable "id"`) {
		t.Errorf("Should not contain computed attribute 'id' in arguments, got: %s", result)
	}

	if strings.Contains(result, `variable "arn"`) {
		t.Errorf("Should not contain computed attribute 'arn' in arguments, got: %s", result)
	}
}

func Test_HCL_Export_Attributes_To_Outputs(t *testing.T) {
	// Set consistent color profile for stable output
	lipgloss.SetColorProfile(0)

	// Load minimal AWS fixture
	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	// Get aws_instance schema directly
	providerSchema := ps.Schemas["registry.terraform.io/hashicorp/aws"]
	instanceSchema := providerSchema.ResourceSchemas["aws_instance"]

	// Test attributes to outputs conversion
	result := ui.ConvertAttributesToHCLOutputs("aws_instance", instanceSchema, "registry.terraform.io/hashicorp/aws")

	// Verify result contains expected output blocks
	if !strings.Contains(result, `output "id"`) {
		t.Errorf("Expected id output, got: %s", result)
	}

	if !strings.Contains(result, `output "arn"`) {
		t.Errorf("Expected arn output, got: %s", result)
	}

	if !strings.Contains(result, `value = aws_instance.instance.id`) {
		t.Errorf("Expected aws_instance reference, got: %s", result)
	}

	if !strings.Contains(result, `value = aws_instance.instance.arn`) {
		t.Errorf("Expected aws_instance arn reference, got: %s", result)
	}

	// Should NOT contain required/optional attributes
	if strings.Contains(result, `output "ami"`) {
		t.Errorf("Should not contain required attribute 'ami' in outputs, got: %s", result)
	}

	if strings.Contains(result, `output "instance_type"`) {
		t.Errorf("Should not contain required attribute 'instance_type' in outputs, got: %s", result)
	}
}

func Test_ConvertToHCL_SwitchBySection(t *testing.T) {
	// Load minimal AWS fixture
	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	// Get aws_instance schema directly
	providerSchema := ps.Schemas["registry.terraform.io/hashicorp/aws"]
	instanceSchema := providerSchema.ResourceSchemas["aws_instance"]

	// Test Arguments section
	argsResult := ui.ConvertToHCL("aws_instance", instanceSchema, ui.ArgumentsSection, "registry.terraform.io/hashicorp/aws")
	
	if !strings.Contains(argsResult, "variable") {
		t.Errorf("Arguments section should produce variables, got: %s", argsResult)
	}

	// Test Attributes section  
	attrsResult := ui.ConvertToHCL("aws_instance", instanceSchema, ui.AttributesSection, "registry.terraform.io/hashicorp/aws")
	
	if !strings.Contains(attrsResult, "output") {
		t.Errorf("Attributes section should produce outputs, got: %s", attrsResult)
	}

	// Test unknown section
	unknownResult := ui.ConvertToHCL("aws_instance", instanceSchema, "Unknown", "registry.terraform.io/hashicorp/aws")
	
	if unknownResult != "# Unknown section type" {
		t.Errorf("Unknown section should return error message, got: %s", unknownResult)
	}
}

func Test_S3_Bucket_Schema(t *testing.T) {
	// Load minimal AWS fixture
	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	// Get aws_s3_bucket schema
	providerSchema := ps.Schemas["registry.terraform.io/hashicorp/aws"]
	bucketSchema := providerSchema.ResourceSchemas["aws_s3_bucket"]

	// Test arguments (bucket is required)
	argsResult := ui.ConvertArgumentsToHCLVariables(bucketSchema)
	if !strings.Contains(argsResult, `variable "bucket"`) {
		t.Errorf("Expected bucket variable, got: %s", argsResult)
	}

	// Test attributes (should have computed fields)
	attrsResult := ui.ConvertAttributesToHCLOutputs("aws_s3_bucket", bucketSchema, "registry.terraform.io/hashicorp/aws")
	if !strings.Contains(attrsResult, `output "bucket_domain_name"`) {
		t.Errorf("Expected bucket_domain_name output, got: %s", attrsResult)
	}

	if !strings.Contains(attrsResult, `value = aws_s3_bucket.bucket.bucket_domain_name`) {
		t.Errorf("Expected proper resource reference, got: %s", attrsResult)
	}
}

func Test_EmptySchema_Handling(t *testing.T) {
	// Create empty schema
	emptySchema := &tfjson.Schema{
		Block: &tfjson.SchemaBlock{
			Attributes:   make(map[string]*tfjson.SchemaAttribute),
			NestedBlocks: make(map[string]*tfjson.SchemaBlockType),
		},
	}

	// Test arguments conversion
	argsResult := ui.ConvertArgumentsToHCLVariables(emptySchema)
	if !strings.Contains(argsResult, "# No arguments available") {
		t.Errorf("Expected no arguments message, got: %s", argsResult)
	}

	// Test attributes conversion
	attrsResult := ui.ConvertAttributesToHCLOutputs("empty", emptySchema, "test")
	if !strings.Contains(attrsResult, "# No computed attributes available") {
		t.Errorf("Expected no attributes message, got: %s", attrsResult)
	}
}