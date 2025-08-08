package ui_test

import (
	"bytes"
	"path/filepath"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/terraconstructs/provider-explorer/internal/ui"
)

func Test_FilterForInstance(t *testing.T) {
	lipgloss.SetColorProfile(0)

	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	m := ui.NewModelWithSchemas(ps, 120, 30)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(120, 30))

	// Wait for initial load 
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		hasProvider := bytes.Contains(b, []byte("registry.terraform.io/hashicorp/aws"))
		hasInstance := bytes.Contains(b, []byte("aws_instance"))
		hasBucket := bytes.Contains(b, []byte("aws_s3_bucket"))
		return hasProvider && hasInstance && hasBucket
	}, teatest.WithDuration(5*time.Second))

	// Try filtering for "instance" instead of "inst"
	tm.Type("/instance")
	
	time.Sleep(200 * time.Millisecond)

	// Check current output
	output := tm.Output()
	buf := make([]byte, 8192)
	n, _ := output.Read(buf)
	result := string(buf[:n])
	
	hasInstance := bytes.Contains(buf[:n], []byte("aws_instance"))
	hasBucket := bytes.Contains(buf[:n], []byte("aws_s3_bucket"))
	
	t.Logf("After filtering for 'instance': Has aws_instance: %v, Has aws_s3_bucket: %v", hasInstance, hasBucket)
	
	if hasInstance && !hasBucket {
		t.Log("✅ Filtering for 'instance' works!")
	} else {
		t.Log("❌ Filtering for 'instance' not working")
		t.Logf("Full output:\n%s", result)
	}

	tm.Quit()
}