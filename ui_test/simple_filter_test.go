package ui_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/terraconstructs/provider-explorer/internal/ui"
)

func Test_SimpleFilterStart(t *testing.T) {
	// Set consistent color profile for stable output
	lipgloss.SetColorProfile(0)

	// Load minimal AWS fixture
	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	// Build the TUI with fixed size for stable output
	m := ui.NewModelWithSchemas(ps, 120, 30)

	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(120, 30),
	)

	// Wait for initial rendering
	time.Sleep(100 * time.Millisecond)

	// Get initial output
	output := tm.Output()
	buf := make([]byte, 4096)
	n, _ := output.Read(buf)
	t.Logf("Initial output:\n%s", string(buf[:n]))

	// Send "/" key
	tm.Type("/")

	// Give it a moment
	time.Sleep(100 * time.Millisecond)

	// Get output after "/" 
	buf2 := make([]byte, 4096)
	n2, _ := output.Read(buf2)
	t.Logf("Output after '/':\n%s", string(buf2[:n2]))

	// Type "bucket"
	tm.Type("bucket")

	// Give it a moment
	time.Sleep(100 * time.Millisecond)

	// Get output after typing "bucket"
	buf3 := make([]byte, 4096)
	n3, _ := output.Read(buf3)
	t.Logf("Output after typing 'bucket':\n%s", string(buf3[:n3]))

	tm.Quit()
}