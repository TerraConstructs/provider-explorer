package ui_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/terraconstructs/provider-explorer/internal/ui"
)

func Test_DebugInstanceFilter(t *testing.T) {
	// Set consistent color profile for stable output
	lipgloss.SetColorProfile(0)

	// Load minimal AWS fixture
	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	// Test "bucket" (known working)
	t.Log("=== Testing 'bucket' filter (known working) ===")
	m1 := ui.NewModelWithSchemas(ps, 120, 30)
	tm1 := teatest.NewTestModel(t, m1, teatest.WithInitialTermSize(120, 30))

	time.Sleep(100 * time.Millisecond)
	tm1.Type("/bucket")
	time.Sleep(100 * time.Millisecond)

	output1 := tm1.Output()
	buf1 := make([]byte, 4096)
	n1, _ := output1.Read(buf1)
	t.Logf("'bucket' filter result:\n%s", string(buf1[:n1]))
	tm1.Quit()

	// Test "inst" (problematic)
	t.Log("=== Testing 'inst' filter (problematic) ===")
	m2 := ui.NewModelWithSchemas(ps, 120, 30)
	tm2 := teatest.NewTestModel(t, m2, teatest.WithInitialTermSize(120, 30))

	time.Sleep(100 * time.Millisecond)
	tm2.Type("/inst")
	time.Sleep(100 * time.Millisecond)

	output2 := tm2.Output()
	buf2 := make([]byte, 4096)
	n2, _ := output2.Read(buf2)
	t.Logf("'inst' filter result:\n%s", string(buf2[:n2]))
	tm2.Quit()

	// Test "aws" (should match both)
	t.Log("=== Testing 'aws' filter (should match both) ===")
	m3 := ui.NewModelWithSchemas(ps, 120, 30)
	tm3 := teatest.NewTestModel(t, m3, teatest.WithInitialTermSize(120, 30))

	time.Sleep(100 * time.Millisecond)
	tm3.Type("/aws")
	time.Sleep(100 * time.Millisecond)

	output3 := tm3.Output()
	buf3 := make([]byte, 4096)
	n3, _ := output3.Read(buf3)
	t.Logf("'aws' filter result:\n%s", string(buf3[:n3]))
	tm3.Quit()
}
