package editor

import (
	"ebmgo/bookmanager"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
)

// PrepareBooksForImport lets users edit book metadata in their preferred editor before importing.
func PrepareBooksForImport(books []bookmanager.Book) error {
	// Create new folder to store a temp file for user to modiefied later
	filePath := filepath.Join(os.TempDir(), "ebm-import.json")

	// Convert to JSON
	jsonData, err := json.MarshalIndent(books, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return err
	}

	// Open the file in an editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}

	cmd := exec.Command(editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Wait user to exit the editor
	if err := cmd.Run(); err != nil {
		return err
	}

	// Get the modified JSON file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(content, &books); err != nil {
		return err
	}

	return nil
}
