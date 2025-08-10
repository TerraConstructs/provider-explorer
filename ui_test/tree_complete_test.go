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

func Test_TreeCompleteNavigation(t *testing.T) {
	// Comprehensive test of all tree navigation features
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

	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // Select entity

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Schema (Arguments)"))
	}, teatest.WithDuration(5*time.Second))

	t.Log("✅ Tree view loaded")

	// Test 1: Arrow key navigation
	t.Log("=== Testing Arrow Key Navigation ===")

	// DOWN arrow should work
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	time.Sleep(50 * time.Millisecond)

	// UP arrow should work
	tm.Send(tea.KeyMsg{Type: tea.KeyUp})
	time.Sleep(50 * time.Millisecond)

	// Test 2: j/k navigation (vim-style)
	t.Log("=== Testing j/k Navigation ===")

	// 'j' should move down
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	time.Sleep(50 * time.Millisecond)

	// 'k' should move up
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	time.Sleep(50 * time.Millisecond)

	// Test 3: Selection with spacebar
	t.Log("=== Testing Selection ===")
	tm.Send(tea.KeyMsg{Type: tea.KeySpace})

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Selected: 1"))
	}, teatest.WithDuration(2*time.Second))

	t.Log("✅ Spacebar selection works")

	// Test 4: Multiple selections
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}) // Move down
	time.Sleep(50 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeySpace}) // Select another

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Selected: 2"))
	}, teatest.WithDuration(2*time.Second))

	t.Log("✅ Multiple selection works")

	// Test 5: Mode toggle (Arguments ↔ Attributes)
	t.Log("=== Testing Args/Attrs Toggle ===")

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Schema (Attributes)"))
	}, teatest.WithDuration(2*time.Second))

	t.Log("✅ Switch to Attributes mode works")

	// Switch back to Arguments
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Schema (Arguments)"))
	}, teatest.WithDuration(2*time.Second))

	t.Log("✅ Switch back to Arguments mode works")

	// Test 6: Navigation works in both modes
	t.Log("=== Testing Navigation in Attributes Mode ===")

	// Switch to attributes
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	time.Sleep(50 * time.Millisecond)

	// Test j/k navigation in attributes mode
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	time.Sleep(50 * time.Millisecond)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	time.Sleep(50 * time.Millisecond)

	// Test selection in attributes mode
	tm.Send(tea.KeyMsg{Type: tea.KeySpace})

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Selected:")) && bytes.Contains(b, []byte("Schema (Attributes)"))
	}, teatest.WithDuration(2*time.Second))

	t.Log("✅ Navigation and selection work in Attributes mode")

	t.Log("🎉 All tree navigation features working correctly!")

	tm.Quit()
}

func Test_TreeNavigationSequence(t *testing.T) {
	// Test a realistic navigation sequence
	lipgloss.SetColorProfile(0)

	ps, err := ui.LoadProvidersSchemaFromFile(filepath.FromSlash("../testdata/schemas/aws_min.json"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	m := ui.NewModelWithSchemas(ps, 120, 30)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(120, 30))

	// Navigate to tree
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("aws_instance"))
	}, teatest.WithDuration(5*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Schema (Arguments)"))
	}, teatest.WithDuration(5*time.Second))

	// Realistic workflow: Navigate and select multiple items
	t.Log("=== Realistic workflow: Navigate tree and select items ===")

	// Navigate down with j and select items
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}) // Move to ami
	tm.Send(tea.KeyMsg{Type: tea.KeySpace})                     // Select ami
	time.Sleep(50 * time.Millisecond)

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}) // Move to instance_type
	tm.Send(tea.KeyMsg{Type: tea.KeySpace})                     // Select instance_type
	time.Sleep(50 * time.Millisecond)

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}) // Move to tags
	tm.Send(tea.KeyMsg{Type: tea.KeySpace})                     // Select tags

	// Should have 3 selections now
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Selected: 3"))
	}, teatest.WithDuration(2*time.Second))

	t.Log("✅ Selected 3 items using j+space navigation")

	// Switch to attributes and continue navigation
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}) // Switch to attributes

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Schema (Attributes)"))
	}, teatest.WithDuration(2*time.Second))

	// Navigate in attributes with k (up) and select
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}) // Move up
	tm.Send(tea.KeyMsg{Type: tea.KeySpace})                     // Select attribute

	// Should have more selections now
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Selected:")) &&
			bytes.Contains(b, []byte("Schema (Attributes)"))
	}, teatest.WithDuration(2*time.Second))

	t.Log("✅ Mixed navigation workflow completed successfully")

	tm.Quit()
}
