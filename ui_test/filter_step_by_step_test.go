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

func Test_FilterStepByStep(t *testing.T) {
	lipgloss.SetColorProfile(0)

	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	m := ui.NewModelWithSchemas(ps, 120, 30)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(120, 30))

	// Step 1: Wait for initial load
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		hasProvider := bytes.Contains(b, []byte("registry.terraform.io/hashicorp/aws"))
		hasInstance := bytes.Contains(b, []byte("aws_instance"))
		hasBucket := bytes.Contains(b, []byte("aws_s3_bucket"))
		return hasProvider && hasInstance && hasBucket
	}, teatest.WithDuration(5*time.Second))

	t.Log("✅ Initial load complete")

	// Step 2: Try to start filtering with "/"
	t.Log("Attempting to start filtering with '/'")
	tm.Type("/")

	time.Sleep(200 * time.Millisecond)

	// Check current output
	output := tm.Output()
	buf := make([]byte, 8192)
	n, _ := output.Read(buf)
	result := string(buf[:n])

	hasFilterMode := bytes.Contains(buf[:n], []byte("filtering")) || bytes.Contains(buf[:n], []byte("Filter:"))
	t.Logf("After typing '/': Has filter mode: %v", hasFilterMode)

	if !hasFilterMode {
		t.Logf("Output after '/':\n%s", result)
		t.Log("❌ Filter mode not activated")
		tm.Quit()
		return
	}

	t.Log("✅ Filter mode activated")

	// Step 3: Type "inst"
	t.Log("Typing 'inst'")
	tm.Type("inst")

	time.Sleep(200 * time.Millisecond)

	// Check current output again
	n, _ = output.Read(buf)
	result = string(buf[:n])

	hasInstText := bytes.Contains(buf[:n], []byte("inst"))
	hasInstance := bytes.Contains(buf[:n], []byte("aws_instance"))
	hasBucket := bytes.Contains(buf[:n], []byte("aws_s3_bucket"))

	t.Logf("After typing 'inst': Has 'inst': %v, Has aws_instance: %v, Has aws_s3_bucket: %v", hasInstText, hasInstance, hasBucket)

	if hasInstText && hasInstance && !hasBucket {
		t.Log("✅ Filtering works correctly!")
	} else {
		t.Log("❌ Filtering not working as expected")
		t.Logf("Final output:\n%s", result)
	}

	tm.Quit()
}
