package bookfinder

import (
	"ebmgo/bookmanager"
	"ebmgo/bookparser"
	"errors"
	"os"
	"path/filepath"
)

var (
	errNotSupportMimeType = errors.New("unsupported mime type")
)

// GetEbooks return books from the path.
// If path is directory getEbook return all supported books.
// If path is file getEbook return books or return error if filetype not supported.
func GetEbooks(path string) ([]bookmanager.Book, error) {
	info, err := os.Stat(path)
	if err != nil {
		return []bookmanager.Book{}, err
	}

	// Path is a file then filetype should support
	// if not return error
	if !info.IsDir() {
		bookInfo, err := bookparser.Parse(path)
		if err != nil {
			return []bookmanager.Book{}, err
		}

		book := bookmanager.NewBook(
			bookInfo.Metadata.ISBN,
			bookInfo.Metadata.Title,
			bookInfo.Metadata.Authors,
			bookInfo.Metadata.Publisher,
			bookInfo.Metadata.Tags,
		)
		book.AppendFiles(bookInfo.File.Path, bookInfo.File.Type)
		return []bookmanager.Book{book}, nil
	}

	// Path is a dir then list files in this directory
	f, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return []bookmanager.Book{}, err
	}
	defer f.Close()

	finfo, err := f.Readdir(0)
	if err != nil {
		return []bookmanager.Book{}, err
	}

	var books []bookmanager.Book
	mapTitle := make(map[string]int)
	i := 0
	for _, item := range finfo {
		// Skip subs directory
		if item.IsDir() {
			continue
		}

		// Insert files if filetype support
		// if not skip
		filePath := filepath.Join(path, item.Name())
		f, err := bookparser.Parse(filePath)
		if err != nil {
			if err == bookparser.ErrNotSupportMimeType {
				continue
			}
			return []bookmanager.Book{}, err
		}

		j, ok := mapTitle[f.Metadata.Title]
		if ok {
			books[j].AppendFiles(f.File.Path, f.File.Type)
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
		i++
	}

	return books, nil
}
