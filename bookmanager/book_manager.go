package bookmanager

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Book is a single/unique book identity. It has many bookfiles to store different book format.
type Book struct {
	ID           int
	ISBN         string
	Title        string
	Authors      []string
	uniqueAuthor map[string]bool
	Publisher    string
	Tags         []string
	uniqueTag    map[string]bool
	BookFiles    []BookFiles
	uniqueFile   map[string]bool
}

// AppendFiles appends file to books.
func (b *Book) AppendFiles(filePath string, fileType string) {
	if !b.uniqueFile[filePath] {
		b.BookFiles = append(b.BookFiles, BookFiles{FilePath: filePath, FileType: fileType})
		b.uniqueFile[filePath] = true
	}
}

func (b *Book) AppendAuthors(author string) {
	if !b.uniqueAuthor[author] {
		b.Authors = append(b.Authors, author)
		b.uniqueAuthor[author] = true
	}
}

func (b *Book) AppendTag(tag string) {
	if !b.uniqueAuthor[tag] {
		b.Tags = append(b.Tags, tag)
		b.uniqueTag[tag] = true
	}
}

// BookFiles is a book file with specific filepath and filetype.
type BookFiles struct {
	FilePath string
	FileType string
}

// NewBook return new instance of book.
func NewBook(isbn string, title string, authors []string, publisher string, tags []string) Book {
	b := Book{
		ISBN:         isbn,
		Title:        title,
		Authors:      []string{},
		uniqueAuthor: make(map[string]bool),
		Publisher:    publisher,
		Tags:         []string{},
		uniqueTag:    make(map[string]bool),
		BookFiles:    []BookFiles{},
		uniqueFile:   make(map[string]bool),
	}

	for _, author := range authors {
		b.AppendAuthors(author)
	}

	for _, tag := range tags {
		b.AppendTag(tag)
	}

	return b
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

type processBookResult struct {
	book *Book
	err  error
}

// processBookToEBMDir copies books from the source directory to the EBM directory.
func (b *BookManager) processBookToEBMDir(book *Book, result chan<- processBookResult) {
	// create folder to store a book
	authors := "Unknown"
	if len(book.Authors) > 0 {
		authors = strings.Join(book.Authors, ",")
	}
	path := filepath.Join(b.directory, authors, book.Title)
	if err := os.MkdirAll(path, 0750); err != nil {
		result <- processBookResult{
			book: nil,
			err:  err,
		}
		return
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
			result <- processBookResult{
				book: nil,
				err:  err,
			}
			return
		}
		defer srcFile.Close()
		destFile, err := os.Create(destPath) // Destination to paste
		if err != nil {
			result <- processBookResult{
				book: nil,
				err:  err,
			}
			return
		}
		defer destFile.Close()
		if _, err := io.Copy(destFile, srcFile); err != nil {
			result <- processBookResult{
				book: nil,
				err:  err,
			}
			return
		}
		book.BookFiles[j].FilePath = destPath
		newBook.AppendFiles(book.BookFiles[j].FilePath, book.BookFiles[j].FileType)
	}
	if len(newBook.BookFiles) > 0 {
		// insert bookfiles
		result <- processBookResult{
			book: &newBook,
			err:  nil,
		}
	}
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
func (b *BookManager) ImportBooks(ctx context.Context, worker int, books []Book) error {
	insertBook := []*Book{} // metadata to store
	if err := b.repo.CreateBooks(
		ctx,
		func() ([]*Book, error) {

			jobs := make(chan *Book)
			result := make(chan processBookResult)
			wg := sync.WaitGroup{}

			for i := 0; i < worker; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for {
						select {
						case <-ctx.Done():
							return
						case book, ok := <-jobs:
							if !ok {
								return
							}
							b.processBookToEBMDir(book, result)
						}
					}
				}()
			}

			go func() {
				for _, book := range books {
					select {
					case <-ctx.Done():
						close(jobs)
						return
					case jobs <- &book:
					}
				}
				close(jobs)
			}()

			go func() {
				wg.Wait()
				close(result)
			}()

			var err error
			for res := range result {
				if res.err != nil {
					err = res.err
				}

				insertBook = append(insertBook, res.book)
			}

			return insertBook, err
		},
		func() {
			for _, book := range insertBook {
				for _, file := range book.BookFiles {
					os.Remove(file.FilePath)
				}
			}
		},
	); err != nil {
		return err
	}

	return nil
}

func getFilename(filepath string) string {
	filename := ""
	for i := len(filepath) - 1; i > 0; i-- {
		char := filepath[i]
		switch char {
		case '/':
			return filename
		default:
			filename = string(char) + filename
		}
	}

	return ""
}

func (h *BookManager) Export(ids []int, dstPath string) error {
	books, err := h.repo.getBooks(ids)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dstPath, 0750); err != nil {
		return err
	}

	for _, book := range books {
		for _, file := range book.BookFiles {
			dst := filepath.Join(dstPath, getFilename(file.FilePath))
			if err := copyFileContents(file.FilePath, dst); err != nil {
				fmt.Println("copy file dst to", dst, "error:", err)
			}
		}
	}

	return nil
}

func (b *BookManager) GetBooks(pattern string) ([]Book, error) {
	return b.repo.FindBooks(pattern)
}

func (b *BookManager) RemoveBooks(ids []int) error {
	var tmpFile []string
	if err := b.repo.RemoveBooks(
		ids,
		func(books []Book) error {
			for _, book := range books {
				for _, file := range book.BookFiles {
					// move files to /tmp directory
					newPath := strings.ReplaceAll(file.FilePath, b.directory, os.TempDir())
					err := moveFile(file.FilePath, newPath)
					if err != nil {
						return fmt.Errorf("move failed: %v", err)
					}
					tmpFile = append(tmpFile, newPath)
				}
			}
			return nil
		},
		func() {
			for _, f := range tmpFile {
				newPath := strings.ReplaceAll(f, os.TempDir(), b.directory)
				copyFileContents(f, newPath)
			}
		},
	); err != nil {
		return err
	}

	for _, f := range tmpFile {
		os.Remove(f)
	}

	return nil
}

func moveFile(src, dst string) error {
	// Ensure parent directories exist
	parentDir := filepath.Dir(dst)
	err := os.MkdirAll(parentDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create parent directory: %v", err)
	}

	err = copyFileContents(src, dst)
	if err != nil {
		return fmt.Errorf("copy failed: %v", err)
	}

	// Remove original file
	err = os.Remove(src)
	if err != nil {
		return fmt.Errorf("remove original failed: %v", err)
	}
	return nil

}

func copyFileContents(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Optional: Copy file permissions
	info, err := sourceFile.Stat()
	if err == nil {
		err = os.Chmod(dst, info.Mode())
	}

	return err
}
