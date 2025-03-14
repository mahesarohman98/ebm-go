package bookfinder

import (
	"ebmgo/bookparser"
	"errors"
	"os"
	"path/filepath"
)

var (
	errNotSupportMimeType = errors.New("unsupported mime type")
)

// GetEbooks return books parser from the path.
// If path is directory getEbook return all supported books.
// If path is file getEbook return ebook or return error if filetype not supported.
func GetEbooks(path string) ([]bookparser.BookParser, error) {
	info, err := os.Stat(path)
	if err != nil {
		return []bookparser.BookParser{}, err
	}

	// Path is a file then filetype should support
	// if not return error
	if !info.IsDir() {
		bookInfo, err := bookparser.Parse(path)
		if err != nil {
			return []bookparser.BookParser{}, err
		}

		return []bookparser.BookParser{bookInfo}, nil
	}

	// Path is a dir then list files in this directory
	f, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return []bookparser.BookParser{}, err
	}
	defer f.Close()

	finfo, err := f.Readdir(0)
	if err != nil {
		return []bookparser.BookParser{}, err
	}

	files := []bookparser.BookParser{}
	for _, item := range finfo {
		// Skip subs directory
		if item.IsDir() {
			continue
		}

		// Insert files if filetype support
		// if not skip
		filePath := filepath.Join(path, item.Name())
		bookInfo, err := bookparser.Parse(filePath)
		if err != nil {
			if err == bookparser.ErrNotSupportMimeType {
				continue
			}
			return []bookparser.BookParser{}, err
		}

		files = append(files, bookInfo)
	}

	return files, nil
}
