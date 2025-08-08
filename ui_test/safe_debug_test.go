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

func Test_SafeDebugWaitCondition(t *testing.T) {
	lipgloss.SetColorProfile(0)

	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	m := ui.NewModelWithSchemas(ps, 120, 30)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(120, 30))

	// Test the wait condition with timeout
	success := false
	timeout := time.NewTimer(3 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)

	defer timeout.Stop()
	defer ticker.Stop()

	for !success {
		select {
		case <-timeout.C:
			t.Log("TIMEOUT: Wait condition not met")
			output := tm.Output()
			buf := make([]byte, 8192)
			n, _ := output.Read(buf)
			t.Logf("Final output:\n%s", string(buf[:n]))
			
			hasProvider := bytes.Contains(buf[:n], []byte("registry.terraform.io/hashicorp/aws"))
			hasInstance := bytes.Contains(buf[:n], []byte("aws_instance"))
			hasBucket := bytes.Contains(buf[:n], []byte("aws_s3_bucket"))
			
			t.Logf("hasProvider: %v", hasProvider)
			t.Logf("hasInstance: %v", hasInstance)
			t.Logf("hasBucket: %v", hasBucket)
			return
			
		case <-ticker.C:
			output := tm.Output()
			buf := make([]byte, 8192)
			n, _ := output.Read(buf)
			
			hasProvider := bytes.Contains(buf[:n], []byte("registry.terraform.io/hashicorp/aws"))
			hasInstance := bytes.Contains(buf[:n], []byte("aws_instance"))
			hasBucket := bytes.Contains(buf[:n], []byte("aws_s3_bucket"))
			
			if hasProvider && hasInstance && hasBucket {
				t.Log("SUCCESS: All conditions met")
				success = true
				break
			}
		}
	}

	tm.Quit()
}