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

func Test_FilterExactTerms(t *testing.T) {
	lipgloss.SetColorProfile(0)

	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	m := ui.NewModelWithSchemas(ps, 120, 30)

	// Test different filter terms
	testCases := []struct {
		filter   string
		expectInstance bool
		expectBucket   bool
	}{
		{"aws", true, true},
		{"instance", true, false},
		{"s3", false, true},
		{"bucket", false, true},
		{"aws_", true, true},
		{"_instance", true, false},
		{"_s3_", false, true},
	}

	for _, tc := range testCases {
		t.Run("filter_"+tc.filter, func(t *testing.T) {
			tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(120, 30))

			// Wait for load
			teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
				return bytes.Contains(b, []byte("aws_instance"))
			}, teatest.WithDuration(2*time.Second))

			// Apply filter
			tm.Type("/" + tc.filter)
			tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
			time.Sleep(100 * time.Millisecond)

			// Check results
			output := tm.Output()
			buf := make([]byte, 8192)
			n, _ := output.Read(buf)
			
			hasInstance := bytes.Contains(buf[:n], []byte("aws_instance"))
			hasBucket := bytes.Contains(buf[:n], []byte("aws_s3_bucket"))
			
			if hasInstance == tc.expectInstance && hasBucket == tc.expectBucket {
				t.Logf("✅ Filter '%s': aws_instance=%v, aws_s3_bucket=%v (expected)", tc.filter, hasInstance, hasBucket)
			} else {
				t.Errorf("❌ Filter '%s': aws_instance=%v (exp %v), aws_s3_bucket=%v (exp %v)", 
					tc.filter, hasInstance, tc.expectInstance, hasBucket, tc.expectBucket)
			}

			tm.Quit()
		})
	}
}