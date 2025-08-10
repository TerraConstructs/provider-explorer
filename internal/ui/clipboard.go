package ui

import (
	"github.com/atotto/clipboard"
)

// CopyToClipboard copies the given content to the system clipboard
func CopyToClipboard(content string) error {
	return clipboard.WriteAll(content)
}

// IsClipboardAvailable checks if clipboard functionality is available
func IsClipboardAvailable() bool {
	// Try to read from clipboard to test availability
	_, err := clipboard.ReadAll()
	// If there's no error or it's just empty, clipboard is available
	// Some errors (like no clipboard daemon) indicate unavailability
	return err == nil || err.Error() == "exit status 1" // xclip returns this for empty clipboard
}
