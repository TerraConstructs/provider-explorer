package ui_test

import (
	"bytes"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/terraconstructs/provider-explorer/internal/ui"
)

func Test_FilterWithEnterToApply(t *testing.T) {
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

	// Start filtering
	tm.Type("/inst")
	time.Sleep(100 * time.Millisecond)
	
	// Press Enter to apply the filter
	t.Log("Pressing Enter to apply filter")
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(200 * time.Millisecond)

	// Check results
	output := tm.Output()
	buf := make([]byte, 8192)
	n, _ := output.Read(buf)
	result := string(buf[:n])
	
	hasInstance := bytes.Contains(buf[:n], []byte("aws_instance"))
	hasBucket := bytes.Contains(buf[:n], []byte("aws_s3_bucket"))
	hasFilteringText := bytes.Contains(buf[:n], []byte("filtering"))
	
	t.Logf("After Enter - Has aws_instance: %v, Has aws_s3_bucket: %v, Has 'filtering': %v", hasInstance, hasBucket, hasFilteringText)
	
	if hasInstance && !hasBucket && !hasFilteringText {
		t.Log("✅ Filter applied successfully!")
	} else {
		t.Log("❌ Filter not applied correctly")
		t.Logf("Full output:\n%s", result)
	}

	tm.Quit()
}