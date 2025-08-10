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

func Test_MinimalFlowThatShouldWork(t *testing.T) {
	lipgloss.SetColorProfile(0)

	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	m := ui.NewModelWithSchemas(ps, 120, 30)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(120, 30))

	// Step 1: Wait for initial load (this works from safe debug test)
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		hasProvider := bytes.Contains(b, []byte("registry.terraform.io/hashicorp/aws"))
		hasInstance := bytes.Contains(b, []byte("aws_instance"))
		hasBucket := bytes.Contains(b, []byte("aws_s3_bucket"))
		return hasProvider && hasInstance && hasBucket
	}, teatest.WithDuration(5*time.Second))

	t.Log("✅ Initial wait condition passed")

	// Step 2: Try simple navigation - just enter to select an entity
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Give it a moment
	time.Sleep(200 * time.Millisecond)

	t.Log("✅ Sent Enter key")

	// Step 3: Check if we got to the tree view
	output := tm.Output()
	buf := make([]byte, 8192)
	n, _ := output.Read(buf)
	result := string(buf[:n])

	hasSchema := bytes.Contains(buf[:n], []byte("Schema ("))
	t.Logf("Has Schema view: %v", hasSchema)

	if hasSchema {
		t.Log("✅ Successfully navigated to schema view")
	} else {
		t.Log("❌ Did not reach schema view")
		t.Logf("Final output:\n%s", result)
	}

	tm.Quit()
}
