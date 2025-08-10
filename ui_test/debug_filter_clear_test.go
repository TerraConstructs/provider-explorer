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

func Test_DebugFilterClear(t *testing.T) {
	lipgloss.SetColorProfile(0)

	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	m := ui.NewModelWithSchemas(ps, 120, 30)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(120, 30))

	// Wait for basic load
	timeout := time.NewTimer(5 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer timeout.Stop()
	defer ticker.Stop()

	for {
		select {
		case <-timeout.C:
			t.Log("TIMEOUT: Basic load not completed")
			output := tm.Output()
			buf := make([]byte, 8192)
			n, _ := output.Read(buf)
			result := string(buf[:n])
			t.Logf("Final output:\n%s", result)

			hasProvider := bytes.Contains(buf[:n], []byte("registry.terraform.io/hashicorp/aws"))
			hasInstance := bytes.Contains(buf[:n], []byte("aws_instance"))
			hasBucket := bytes.Contains(buf[:n], []byte("aws_s3_bucket"))
			hasFilterHint := bytes.Contains(buf[:n], []byte("press / to filter"))

			t.Logf("hasProvider: %v", hasProvider)
			t.Logf("hasInstance: %v", hasInstance)
			t.Logf("hasBucket: %v", hasBucket)
			t.Logf("hasFilterHint: %v", hasFilterHint)
			tm.Quit()
			return

		case <-ticker.C:
			output := tm.Output()
			buf := make([]byte, 8192)
			n, _ := output.Read(buf)

			hasProvider := bytes.Contains(buf[:n], []byte("registry.terraform.io/hashicorp/aws"))
			hasInstance := bytes.Contains(buf[:n], []byte("aws_instance"))
			hasBucket := bytes.Contains(buf[:n], []byte("aws_s3_bucket"))
			hasFilterHint := bytes.Contains(buf[:n], []byte("press / to filter"))

			if hasProvider && hasInstance && hasBucket && hasFilterHint {
				t.Log("SUCCESS: All conditions met")
				tm.Quit()
				return
			}

			// Log partial progress
			if hasProvider {
				t.Logf("Progress: provider=%v instance=%v bucket=%v filter=%v", hasProvider, hasInstance, hasBucket, hasFilterHint)
			}
		}
	}
}
