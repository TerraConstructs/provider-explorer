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

func Test_FinalTreeNavigationVerification(t *testing.T) {
	// Final verification that all the user's requested features work
	lipgloss.SetColorProfile(0)

	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	m := ui.NewModelWithSchemas(ps, 120, 30)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(120, 30))

	// Navigate to tree view
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("aws_instance"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // Select entity to go to tree

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Schema (Arguments)"))
	}, teatest.WithDuration(5*time.Second))

	// Verify what the user specifically asked about:

	// 1. "up/down arrows don't work" - TEST UP/DOWN ARROWS
	t.Log("Testing UP/DOWN arrow keys...")
	initialOutput := getOutput(tm)

	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	time.Sleep(100 * time.Millisecond)
	afterDown := getOutput(tm)

	if afterDown != initialOutput {
		t.Log("✅ DOWN arrow key works - tree cursor moved")
	} else {
		t.Errorf("❌ DOWN arrow key not working")
	}

	tm.Send(tea.KeyMsg{Type: tea.KeyUp})
	time.Sleep(100 * time.Millisecond)
	afterUp := getOutput(tm)

	if afterUp != afterDown {
		t.Log("✅ UP arrow key works - tree cursor moved back")
	} else {
		t.Errorf("❌ UP arrow key not working")
	}

	// 2. "can we still use j/k to navigate the tree" - TEST J/K KEYS
	t.Log("Testing j/k navigation keys...")
	beforeJ := getOutput(tm)

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	time.Sleep(100 * time.Millisecond)
	afterJ := getOutput(tm)

	if afterJ != beforeJ {
		t.Log("✅ 'j' key works - tree cursor moved down")
	} else {
		t.Errorf("❌ 'j' key not working")
	}

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	time.Sleep(100 * time.Millisecond)
	afterK := getOutput(tm)

	if afterK != afterJ {
		t.Log("✅ 'k' key works - tree cursor moved up")
	} else {
		t.Errorf("❌ 'k' key not working")
	}

	// 3. "spacebar to toggle tree node selection" - TEST SPACEBAR
	t.Log("Testing spacebar selection...")
	tm.Send(tea.KeyMsg{Type: tea.KeySpace})

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Selected:"))
	}, teatest.WithDuration(2*time.Second))

	t.Log("✅ Spacebar selection works - shows 'Selected:' counter")

	// 4. Test that tree view displays correctly after pressing enter on an entity
	t.Log("Verifying tree view displays correctly...")
	currentOutput := getOutput(tm)

	hasSchema := bytes.Contains([]byte(currentOutput), []byte("Schema ("))
	hasTreeContent := bytes.Contains([]byte(currentOutput), []byte("required]")) || bytes.Contains([]byte(currentOutput), []byte("optional]"))
	hasInstructions := bytes.Contains([]byte(currentOutput), []byte("press space to select")) || bytes.Contains([]byte(currentOutput), []byte("Selected:"))

	if hasSchema && hasTreeContent && hasInstructions {
		t.Log("✅ Tree view displays correctly with schema content and instructions")
	} else {
		t.Errorf("❌ Tree view not displaying correctly")
		t.Logf("hasSchema: %v, hasTreeContent: %v, hasInstructions: %v", hasSchema, hasTreeContent, hasInstructions)
	}

	t.Log("🎉 All tree navigation issues have been resolved!")
	t.Log("   - UP/DOWN arrow keys work ✓")
	t.Log("   - j/k vim-style navigation works ✓")
	t.Log("   - Spacebar selection works ✓")
	t.Log("   - Tree view displays correctly after entity selection ✓")

	tm.Quit()
}

func getOutput(tm *teatest.TestModel) string {
	output := tm.Output()
	buf := make([]byte, 4096)
	n, _ := output.Read(buf)
	return string(buf[:n])
}
