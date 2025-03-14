package editor

import (
	"ebmgo/bookmanager"
	"ebmgo/bookparser"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"
)

// PrepareBooksForImport lets users edit book metadata in their preferred editor before importing.
func PrepareBooksForImport(bookInfos []bookparser.BookParser) ([]bookmanager.Book, error) {
	// Create new folder to store a temp file for user to modiefied later
	cd := filepath.Join(os.TempDir(), "ebmgo", uuid.NewString())
	err := os.MkdirAll(cd, 0750)
	if err != nil {
		return []bookmanager.Book{}, err
	}
	filePath := filepath.Join(cd, "data.json")

	// Books to edit
	books := []bookmanager.Book{}
	mapTitle := make(map[string]int)
	for _, f := range bookInfos {
		i, ok := mapTitle[f.Metadata.Title]
		if ok {
			books[i].AppendFiles(f.File.Path, f.File.Type)
		} else {
			book := bookmanager.NewBook(
				f.Metadata.ISBN,
				f.Metadata.Title,
				f.Metadata.Authors,
				f.Metadata.Publisher,
				f.Metadata.Tags,
			)
			book.AppendFiles(f.File.Path, f.File.Type)
			books = append(books, book)
			mapTitle[f.Metadata.Title] = len(books) - 1
		}
	}
	// Convert to JSON
	jsonData, err := json.MarshalIndent(books, "", "  ")
	if err != nil {
		return []bookmanager.Book{}, err
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
			return []bookmanager.Book{}, err
		}
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
		return []bookmanager.Book{}, err
	}

	// Get the modified JSON file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return []bookmanager.Book{}, err
	}

	if err := json.Unmarshal(content, &books); err != nil {
		return []bookmanager.Book{}, err
	}

	return books, nil
}
