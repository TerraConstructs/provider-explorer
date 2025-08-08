package ui_test

import (
	"path/filepath"
	"testing"

	"github.com/terraconstructs/provider-explorer/internal/ui"
)

func Test_LoadFixture(t *testing.T) {
	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	if len(ps.Schemas) == 0 {
		t.Fatal("No provider schemas loaded")
	}

	awsProvider, exists := ps.Schemas["registry.terraform.io/hashicorp/aws"]
	if !exists {
		t.Fatal("AWS provider not found")
	}

	t.Logf("Resource count: %d", len(awsProvider.ResourceSchemas))
	t.Logf("Data source count: %d", len(awsProvider.DataSourceSchemas))

	if len(awsProvider.ResourceSchemas) == 0 {
		t.Fatal("No resources found")
	}

	// Check specific resources
	if _, exists := awsProvider.ResourceSchemas["aws_instance"]; !exists {
		t.Fatal("aws_instance not found")
	}

	if _, exists := awsProvider.ResourceSchemas["aws_s3_bucket"]; !exists {
		t.Fatal("aws_s3_bucket not found")
	}
}

func Test_ModelWithSchemas(t *testing.T) {
	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	m := ui.NewModelWithSchemas(ps, 120, 30)
	
	// Check if schemas are set
	if m.GetSchemas() == nil {
		t.Fatal("Schemas not set in model")
	}

	if len(m.GetSchemas().Schemas) == 0 {
		t.Fatal("No schemas in model")
	}
}