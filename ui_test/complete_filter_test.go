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

func Test_CompleteFilterWorkflow(t *testing.T) {
	// Set consistent color profile for stable output
	lipgloss.SetColorProfile(0)

	// Load minimal AWS fixture (has aws_instance, aws_s3_bucket)
	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	// Build the TUI with fixed size for stable output
	m := ui.NewModelWithSchemas(ps, 120, 30)

	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(120, 30),
	)

	// 1. Wait for initial state - should show both resources
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("aws_instance")) && 
		       bytes.Contains(b, []byte("aws_s3_bucket")) &&
		       bytes.Contains(b, []byte("press / to filter"))
	}, teatest.WithDuration(5*time.Second))

	t.Log("✅ Initial state: Both resources visible")

	// 2. Start filtering with "/"
	tm.Type("/")

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Filter:")) && 
		       bytes.Contains(b, []byte("Enter=apply, Esc=cancel"))
	}, teatest.WithDuration(5*time.Second))

	t.Log("✅ Filter mode activated: Filter input visible with instructions")

	// 3. Type "inst" to filter for aws_instance
	tm.Type("inst")

	time.Sleep(100 * time.Millisecond)
	output := tm.Output()
	buf := make([]byte, 4096)
	n, _ := output.Read(buf)

	if !bytes.Contains(buf[:n], []byte("aws_instance")) {
		t.Errorf("aws_instance should be visible after filtering for 'inst'")
	}
	if bytes.Contains(buf[:n], []byte("aws_s3_bucket")) {
		t.Errorf("aws_s3_bucket should be hidden after filtering for 'inst'")
	}

	t.Log("✅ Filtering works: Only aws_instance visible")

	// 4. Test Enter to apply filter
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	time.Sleep(100 * time.Millisecond)
	output = tm.Output()
	n, _ = output.Read(buf)

	if !bytes.Contains(buf[:n], []byte("aws_instance")) {
		t.Errorf("aws_instance should still be visible after applying filter")
	}
	if bytes.Contains(buf[:n], []byte("Filter:")) {
		t.Errorf("Filter input should be hidden after applying filter")
	}

	t.Log("✅ Enter applies filter: Input hidden, filtered results remain")

	// 5. Test Escape to clear applied filter
	tm.Send(tea.KeyMsg{Type: tea.KeyEscape})

	time.Sleep(100 * time.Millisecond)
	output = tm.Output()
	n, _ = output.Read(buf)

	if !bytes.Contains(buf[:n], []byte("aws_instance")) || !bytes.Contains(buf[:n], []byte("aws_s3_bucket")) {
		t.Errorf("Both resources should be visible after clearing applied filter")
	}

	t.Log("✅ Escape clears applied filter: Both resources visible again")

	// 6. Test cancel during filtering (Escape while typing)
	tm.Type("/") // Start filtering again
	tm.Type("bucket")

	time.Sleep(50 * time.Millisecond)

	tm.Send(tea.KeyMsg{Type: tea.KeyEscape}) // Cancel while filtering

	time.Sleep(100 * time.Millisecond)
	output = tm.Output()
	n, _ = output.Read(buf)

	if !bytes.Contains(buf[:n], []byte("aws_instance")) || !bytes.Contains(buf[:n], []byte("aws_s3_bucket")) {
		t.Errorf("Both resources should be visible after canceling filter")
	}
	if bytes.Contains(buf[:n], []byte("Filter:")) {
		t.Errorf("Filter input should be hidden after canceling")
	}

	t.Log("✅ Escape cancels filtering: Returns to original list")

	t.Log("🎉 Complete filter workflow test passed!")

	tm.Quit()
}