package bookfinder

import (
	"ebmgo/bookmanager"
	"ebmgo/bookparser"
	"errors"
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
		if err == bookparser.ErrNotSupportMimeType {
			return err
		}
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

func (c *collector) walkDir(recursive bool, path string, jobs chan<- string) error {
	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, item := range files {
		filePath := filepath.Join(path, item.Name())
		if item.IsDir() && recursive {
			if err := c.walkDir(recursive, filePath, jobs); err != nil {
				continue
			}
			continue
		}
		jobs <- filePath
	}

	return nil
}

func (c *collector) getEbooks(recursive bool, path string) error {
	jobs := make(chan string)
	wg := sync.WaitGroup{}

	workerCount := 64
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for filePath := range jobs {
				err := c.addOrAppendBook(filePath)
				if err != nil {
				}
			}
		}(i)
	}

	go func() {
		if err := c.walkDir(recursive, path, jobs); err != nil {
		}
		close(jobs)
	}()

	wg.Wait()

	return nil
}

// GetEbooks return books from the path.
// If path is directory getEbook return all supported books.
// If path is file getEbook return books or return error if filetype not supported.
func GetEbooks(recursive bool, path string) ([]bookmanager.Book, error) {
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
	err = collector.getEbooks(recursive, path)
	if err != nil {
		return []bookmanager.Book{}, err
	}

	return collector.books, nil
}
