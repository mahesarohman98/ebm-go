package bookfinder

import (
	"context"
	"ebmgo/bookmanager"
	"ebmgo/bookparser"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	errNotSupportMimeType = errors.New("unsupported mime type")
)

type collector struct {
	books    []bookmanager.Book
	titleMap map[string]int
	mu       sync.Mutex
}

func newCollector() *collector {
	return &collector{
		books:    []bookmanager.Book{},
		titleMap: make(map[string]int),
	}
}

func (c *collector) addOrAppendBook(path string) error {
	f, err := bookparser.Parse(path)
	if err != nil {
		return err
	}

	c.mu.Lock()
	i, found := c.titleMap[f.Metadata.Title]
	if found {
		c.books[i].AppendFiles(f.File.Path, f.File.Type)
	} else {
		book := bookmanager.NewBook(
			f.Metadata.ISBN,
			f.Metadata.Title,
			f.Metadata.Authors,
			f.Metadata.Publisher,
			f.Metadata.Tags,
		)
		book.AppendFiles(f.File.Path, f.File.Type)
		c.books = append(c.books, book)

		c.titleMap[f.Metadata.Title] = len(c.books) - 1
	}
	c.mu.Unlock()

	return nil
}

func (c *collector) walkDir(ctx context.Context, cancel context.CancelCauseFunc, recursive bool, path string, jobs chan<- string) {
	files, err := os.ReadDir(path)
	if err != nil {
		cancel(err)
		return
	}

	for _, item := range files {
		filePath := filepath.Join(path, item.Name())
		if item.IsDir() && recursive {
			c.walkDir(ctx, cancel, recursive, filePath, jobs)
			continue
		}
		jobs <- filePath
	}
}

func (c *collector) getEbooks(worker int, recursive bool, path string) error {
	ctx, cancel := context.WithCancelCause(context.Background())
	jobs := make(chan string)
	wg := sync.WaitGroup{}

	for i := 0; i < worker; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case filePath, ok := <-jobs:
					if !ok {
						return
					}
					err := c.addOrAppendBook(filePath)
					if err != nil {
						if err != bookparser.ErrNotSupportMimeType {
							cancel(fmt.Errorf("error add or append book %s %v", filePath, err))
						}
					}
				}
			}
		}()
	}

	go func() {
		c.walkDir(ctx, cancel, recursive, path, jobs)
		close(jobs)
	}()

	wg.Wait()

	if ctx.Err() != nil {
		return context.Cause(ctx)
	}

	return nil
}

// GetEbooks return books from the path.
// If path is directory getEbook return all supported books.
// If path is file getEbook return books or return error if filetype not supported.
func GetEbooks(worker int, recursive bool, path string) ([]bookmanager.Book, error) {
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

	collector := newCollector()
	err = collector.getEbooks(worker, recursive, path)
	if err != nil {
		return []bookmanager.Book{}, err
	}

	return collector.books, nil
}
