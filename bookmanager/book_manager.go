package bookmanager

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Book is a single/unique book identity. It has many bookfiles to store different book format.
type Book struct {
	ID        int
	ISBN      string
	Title     string
	Authors   []string
	Publisher string
	Tags      []string
	BookFiles []BookFiles
}

// AppendFiles appends file to books.
func (b *Book) AppendFiles(filePath string, fileType string) {
	b.BookFiles = append(b.BookFiles, BookFiles{FilePath: filePath, FileType: fileType})
}

// BookFiles is a book file with specific filepath and filetype.
type BookFiles struct {
	FilePath string
	FileType string
}

// NewBook return new instance of book.
func NewBook(isbn string, title string, author []string, publisher string, tags []string) Book {
	return Book{
		ISBN:      isbn,
		Title:     title,
		Authors:   author,
		Publisher: publisher,
		Tags:      tags,
		BookFiles: []BookFiles{},
	}
}

// BookManager manage books by storing it to ebm dir and db.
type BookManager struct {
	repo      *repository
	directory string
}

// NewBookManage return instance of book manager.
func NewBookManager(ebmDir string) (*BookManager, error) {
	db, err := newSqliteConnection(ebmDir)
	if err != nil {
		return nil, fmt.Errorf("failed connect sqlite: %v", err)
	}
	repo := newRepository(db)
	return &BookManager{repo: repo, directory: ebmDir}, nil
}

func (b *BookManager) Close() {
	b.repo.db.Close()
}

// ImportBooks copies books from the source directory to the EBM directory
// and inserts metadata into the database.
//
// The EBM directory follows this structure:
//
//	ebm-dir/{book.authors}/{book.title}/{book.title} - {book-author}.{ext}
//
// Parameters:
//
//	books []Book - List of books to be imported.
//
// Returns:
//
//	error - An error if any operation fails.
func (b *BookManager) ImportBooks(books []Book) error {
	insertBook := []Book{} // metadata to store
	if err := b.repo.CreateBooks(func() ([]Book, error) {
		for i, book := range books {
			// create folder to store a book
			authors := "Unknown"
			if len(book.Authors) > 0 {
				authors = strings.Join(book.Authors, ",")
			}
			path := filepath.Join(b.directory, authors, book.Title)
			if err := os.MkdirAll(path, 0750); err != nil {
				return []Book{}, err
			}

			newBook := NewBook(book.ISBN, book.Title, book.Authors, book.Publisher, book.Tags)
			for j, file := range book.BookFiles {
				filename := fmt.Sprintf("%s - %s%s", book.Title, authors, filepath.Ext(file.FilePath))
				destPath := filepath.Join(path, filename)

				// if already exist, skip
				if _, err := os.Stat(destPath); !os.IsNotExist(err) {
					continue
				}
				// Copy and paste file
				// Update books[i].Path
				srcFile, err := os.Open(file.FilePath) // Source file to copy
				if err != nil {
					return []Book{}, err
				}
				defer srcFile.Close()
				destFile, err := os.Create(destPath) // Destination to paste
				if err != nil {
					return []Book{}, err
				}
				defer destFile.Close()
				if _, err := io.Copy(destFile, srcFile); err != nil {
					return []Book{}, err
				}
				books[i].BookFiles[j].FilePath = destPath
				newBook.AppendFiles(books[i].BookFiles[j].FilePath, books[i].BookFiles[j].FileType)
			}
			if len(newBook.BookFiles) > 0 {
				insertBook = append(insertBook, newBook)
			}
		}

		return insertBook, nil

	}); err != nil {
		return err
	}

	return nil
}

func (b *BookManager) GetBooks(pattern string) ([]Book, error) {
	return b.repo.FindBooks(pattern)
}
