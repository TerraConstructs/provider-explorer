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

func Test_TreeNavigationKeysIssue(t *testing.T) {
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

	// Navigate to tree view: entities → select entity → tree

	// 1. Wait for entities list to render
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("aws_instance")) &&
			bytes.Contains(b, []byte("aws_s3_bucket"))
	}, teatest.WithDuration(5*time.Second))

	t.Log("✅ Entities view loaded")

	// 2. Select first entity (aws_instance) to go to tree view
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// 3. Wait for tree view to appear
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Schema (Arguments)")) &&
			bytes.Contains(b, []byte("press space to select"))
	}, teatest.WithDuration(5*time.Second))

	t.Log("✅ Tree view loaded")

	// Get initial tree state
	output := tm.Output()
	buf := make([]byte, 8192)
	n, _ := output.Read(buf)
	initialState := string(buf[:n])
	t.Logf("Initial tree state:\n%s", initialState)

	// 4. Test arrow key navigation (DOWN)
	t.Log("=== Testing DOWN arrow key ===")
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	time.Sleep(100 * time.Millisecond)

	n, _ = output.Read(buf)
	afterDownArrow := string(buf[:n])
	t.Logf("After DOWN arrow:\n%s", afterDownArrow)

	// Check if anything changed
	if afterDownArrow == initialState {
		t.Errorf("❌ DOWN arrow key had no effect on tree")
	} else {
		t.Log("✅ DOWN arrow key changed tree state")
	}

	// 5. Test arrow key navigation (UP)
	t.Log("=== Testing UP arrow key ===")
	tm.Send(tea.KeyMsg{Type: tea.KeyUp})
	time.Sleep(100 * time.Millisecond)

	n, _ = output.Read(buf)
	afterUpArrow := string(buf[:n])
	t.Logf("After UP arrow:\n%s", afterUpArrow)

	// 6. Test j key navigation
	t.Log("=== Testing 'j' key (down) ===")
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	time.Sleep(100 * time.Millisecond)

	n, _ = output.Read(buf)
	afterJKey := string(buf[:n])
	t.Logf("After 'j' key:\n%s", afterJKey)

	// 7. Test k key navigation
	t.Log("=== Testing 'k' key (up) ===")
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	time.Sleep(100 * time.Millisecond)

	n, _ = output.Read(buf)
	afterKKey := string(buf[:n])
	t.Logf("After 'k' key:\n%s", afterKKey)

	// 8. Test spacebar selection
	t.Log("=== Testing spacebar selection ===")
	tm.Send(tea.KeyMsg{Type: tea.KeySpace})
	time.Sleep(100 * time.Millisecond)

	n, _ = output.Read(buf)
	afterSpace := string(buf[:n])
	t.Logf("After spacebar:\n%s", afterSpace)

	// Check if selection counter appeared
	hasSelection := bytes.Contains([]byte(afterSpace), []byte("Selected:"))
	if hasSelection {
		t.Log("✅ Spacebar selection works")
	} else {
		t.Errorf("❌ Spacebar selection had no effect")
	}

	// 9. Test 'a' key for args/attrs toggle
	t.Log("=== Testing 'a' key (args/attrs toggle) ===")
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	time.Sleep(100 * time.Millisecond)

	n, _ = output.Read(buf)
	afterAKey := string(buf[:n])
	t.Logf("After 'a' key:\n%s", afterAKey)

	// Check if mode changed from Arguments to Attributes
	hasAttributes := bytes.Contains([]byte(afterAKey), []byte("Schema (Attributes)"))
	if hasAttributes {
		t.Log("✅ 'a' key toggle works (switched to Attributes)")
	} else if bytes.Contains([]byte(afterAKey), []byte("Schema (Arguments)")) {
		t.Log("⚠️  'a' key - still showing Arguments (might be expected)")
	} else {
		t.Errorf("❌ 'a' key had no visible effect")
	}

	tm.Quit()
}

func Test_TreeFocusTransition(t *testing.T) {
	// Test specifically the focus transition from entities to tree
	lipgloss.SetColorProfile(0)

	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	m := ui.NewModelWithSchemas(ps, 120, 30)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(120, 30))

	// Navigate step by step and check focus at each stage

	// 1. Initial state - entities should be focused
	time.Sleep(100 * time.Millisecond)
	output := tm.Output()
	buf := make([]byte, 4096)
	n, _ := output.Read(buf)
	initialState := string(buf[:n])

	if bytes.Contains([]byte(initialState), []byte("press / to filter")) {
		t.Log("✅ Entities are initially focused (has filter hint)")
	} else {
		t.Log("⚠️  Entities focus state unclear from initial output")
	}

	// 2. Press Enter to select entity and go to tree
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	time.Sleep(200 * time.Millisecond)

	n, _ = output.Read(buf)
	treeState := string(buf[:n])
	t.Logf("After Enter (should be tree view):\n%s", treeState)

	// Check if we're in tree view and if it looks properly focused
	hasTreeTitle := bytes.Contains([]byte(treeState), []byte("Schema (Arguments)"))
	hasTreeHint := bytes.Contains([]byte(treeState), []byte("press space to select"))

	if hasTreeTitle && hasTreeHint {
		t.Log("✅ Successfully transitioned to tree view")
	} else {
		t.Errorf("❌ Failed to transition to tree view properly")
	}

	// 3. Test if tree responds to keys immediately after focus
	tm.Send(tea.KeyMsg{Type: tea.KeySpace}) // Should select current node
	time.Sleep(100 * time.Millisecond)

	n, _ = output.Read(buf)
	afterImmediateSpace := string(buf[:n])

	if bytes.Contains([]byte(afterImmediateSpace), []byte("Selected:")) {
		t.Log("✅ Tree immediately responsive to keys after focus transition")
	} else {
		t.Errorf("❌ Tree not responsive immediately after focus transition")
	}

	tm.Quit()
}
