package bookmanager

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type SqliteBook struct {
	BookID       int
	Title        string
	ISBN         string
	CreateDate   time.Time
	ModifiedDate time.Time
}

func newSqliteBook(
	title string,
	isbn string,
) SqliteBook {
	return SqliteBook{
		Title:        title,
		ISBN:         isbn,
		CreateDate:   time.Now(),
		ModifiedDate: time.Now(),
	}
}

type SqliteBookFile struct {
	BookID       int
	FilePath     string
	FileType     string
	CreateDate   time.Time
	ModifiedDate time.Time
}

func newSqliteBookFile(
	bookID int,
	filePath string,
	fileType string,
) SqliteBookFile {
	return SqliteBookFile{
		BookID:       bookID,
		FilePath:     filePath,
		FileType:     fileType,
		CreateDate:   time.Now(),
		ModifiedDate: time.Now(),
	}
}

type SqliteBookTag struct {
	BookID int
	Tags   []string
}

func newSqliteBookTags(bookID int, tags []string) SqliteBookTag {
	return SqliteBookTag{
		BookID: bookID,
		Tags:   tags,
	}
}

type SqliteBookAuthor struct {
	BookID  int
	Authors []string
}

func newSqliteBookAuthors(bookID int, authors []string) SqliteBookAuthor {
	return SqliteBookAuthor{
		BookID:  bookID,
		Authors: authors,
	}
}

type BookManagerRepository struct {
	db *sqlx.DB
}

func newBookManagerRepository(db *sqlx.DB) *BookManagerRepository {
	return &BookManagerRepository{db}
}

func rollback(tx *sqlx.Tx, books []Book) {
	tx.Rollback()
	for _, book := range books {
		for _, file := range book.BookFiles {
			os.Remove(file.FilePath)
		}
	}
}

func (repo *BookManagerRepository) CreateBooks(
	fn func() ([]Book, error),
) error {
	tx, err := repo.db.Beginx()
	if err != nil {
		return err
	}
	var books []Book
	defer func() {
		if err != nil {
			rollback(tx, books)
		} else {
			tx.Commit()
		}
	}()

	books, err = fn()
	if err != nil {
		return err
	}
	if len(books) <= 0 {
		return nil
	}

	if err = repo.batchInsertBooks(tx, books); err != nil {
		return err
	}

	return nil
}

func (repo *BookManagerRepository) batchInsertTags(tx *sqlx.Tx, bookTags []SqliteBookTag) error {
	bookTagsString := make([]string, 0)
	bookTagsValueArgs := make([]interface{}, 0)
	bookTagsParam := 1

	tagsValueStrings := make([]string, 0)
	tagsValueArgs := make([]interface{}, 0)
	tagsParam := 1
	for _, bookTag := range bookTags {
		for _, tag := range bookTag.Tags {
			if tag == "" {
				continue
			}
			bookTagsString = append(bookTagsString, fmt.Sprintf("($%d, $%d)", bookTagsParam, bookTagsParam+1))
			bookTagsValueArgs = append(bookTagsValueArgs, bookTag.BookID)
			bookTagsValueArgs = append(bookTagsValueArgs, tag)
			bookTagsParam += 2

			tagsValueStrings = append(tagsValueStrings, fmt.Sprintf("($%d)", tagsParam))
			tagsValueArgs = append(tagsValueArgs, tag)
			tagsParam++
		}
	}

	if tagsParam > 1 {
		query := fmt.Sprintf(`
            INSERT INTO Tags (tag) VALUES %s ON CONFLICT(tag) DO NOTHING
            `, strings.Join(tagsValueStrings, ","))
		_, err := tx.Exec(query, tagsValueArgs...)

		if err != nil {
			return err
		}
	}

	if bookTagsParam > 1 {
		query := fmt.Sprintf(`
            INSERT INTO BookTags (bookId, tag) VALUES %s ON CONFLICT(bookId, tag) DO NOTHING
            `, strings.Join(bookTagsString, ","))
		_, err := tx.Exec(query, bookTagsValueArgs...)

		if err != nil {
			return err
		}
	}

	return nil
}

func (repo *BookManagerRepository) batchInsertAuthors(tx *sqlx.Tx, bookAuthors []SqliteBookAuthor) error {
	bookAuthorsString := make([]string, 0)
	bookAuthorsValueArgs := make([]interface{}, 0)
	bookAuthorsParam := 1

	authorsString := make([]string, 0)
	authorsValueArgs := make([]interface{}, 0)
	authorsParam := 1
	for _, bookAuthor := range bookAuthors {
		for _, author := range bookAuthor.Authors {
			if author == "" {
				continue
			}
			bookAuthorsString = append(bookAuthorsString, fmt.Sprintf("($%d, $%d)", bookAuthorsParam, bookAuthorsParam+1))
			bookAuthorsValueArgs = append(bookAuthorsValueArgs, bookAuthor.BookID)
			bookAuthorsValueArgs = append(bookAuthorsValueArgs, author)
			bookAuthorsParam += 2

			authorsString = append(authorsString, fmt.Sprintf("($%d)", authorsParam))
			authorsValueArgs = append(authorsValueArgs, author)
			authorsParam++
		}
	}

	if authorsParam > 1 {
		query := fmt.Sprintf(`
            INSERT INTO Authors (author) VALUES %s ON CONFLICT(author) DO NOTHING
            `, strings.Join(authorsString, ","))
		_, err := tx.Exec(query, authorsValueArgs...)
		if err != nil {
			return err
		}
	}

	if bookAuthorsParam > 1 {
		query := fmt.Sprintf(`
            INSERT INTO BookAuthors (bookId, author) VALUES %s ON CONFLICT(bookId, author) DO NOTHING
            `, strings.Join(bookAuthorsString, ","))
		_, err := tx.Exec(query, bookAuthorsValueArgs...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (repo *BookManagerRepository) batchInsertBooks(tx *sqlx.Tx, books []Book) error {
	booksDB := make([]SqliteBook, len(books))
	valueStrings := make([]string, 0)
	valueArgs := make([]interface{}, 0)
	bookParam := 1
	for i, book := range books {
		booksDB[i] = newSqliteBook(book.Title, book.ISBN)
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", bookParam, bookParam+1, bookParam+2, bookParam+3, bookParam+4))
		valueArgs = append(valueArgs, nil)
		valueArgs = append(valueArgs, booksDB[i].Title)
		valueArgs = append(valueArgs, booksDB[i].ISBN)
		valueArgs = append(valueArgs, booksDB[i].CreateDate)
		valueArgs = append(valueArgs, booksDB[i].ModifiedDate)
		bookParam += 5
	}

	if bookParam <= 1 {
		return nil
	}

	query := fmt.Sprintf(`
        INSERT INTO Books (bookId, title, isbn, createDate, modifiedDate) VALUES %s
	`, strings.Join(valueStrings, ","))
	res, err := tx.Exec(query, valueArgs...)
	if err != nil {
		return err
	}
	// Get the last inserted ID
	lastID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// Assign IDs to booksDB
	// Setup batch insert bookFiles
	startID := int(lastID) - len(booksDB) + 1
	bookFileValueString := make([]string, 0)
	bookFileValueArgs := make([]interface{}, 0)
	bookFileParam := 1
	for i := range booksDB {
		booksDB[i].BookID = startID + i
		for _, file := range books[i].BookFiles {
			bookFile := newSqliteBookFile(booksDB[i].BookID, file.FileType, file.FilePath)
			bookFileValueString = append(bookFileValueString, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", bookFileParam, bookFileParam+1, bookFileParam+2, bookFileParam+3, bookFileParam+4))
			bookFileValueArgs = append(bookFileValueArgs, bookFile.BookID)
			bookFileValueArgs = append(bookFileValueArgs, bookFile.FilePath)
			bookFileValueArgs = append(bookFileValueArgs, bookFile.FileType)
			bookFileValueArgs = append(bookFileValueArgs, bookFile.CreateDate)
			bookFileValueArgs = append(bookFileValueArgs, bookFile.ModifiedDate)
			bookFileParam += 5
		}
	}

	query = fmt.Sprintf(`
        INSERT INTO BookFiles (bookId, filePath, fileType, createDate, modifiedDate) VALUES %s
        `, strings.Join(bookFileValueString, ","))
	_, err = tx.Exec(query, bookFileValueArgs...)
	if err != nil {
		return err
	}

	bookTags := []SqliteBookTag{}
	bookAuthors := []SqliteBookAuthor{}
	for i, book := range books {
		if len(book.Tags) > 0 {
			bookTags = append(bookTags, newSqliteBookTags(booksDB[i].BookID, book.Tags))
		}
		if len(book.Authors) > 0 {
			bookAuthors = append(bookAuthors, newSqliteBookAuthors(booksDB[i].BookID, book.Authors))
		}
	}

	if err := repo.batchInsertAuthors(tx, bookAuthors); err != nil {
		return fmt.Errorf("error when batch insert authors: %v", err)
	}

	if err := repo.batchInsertTags(tx, bookTags); err != nil {
		return fmt.Errorf("error when batch insert tags: %v", err)
	}

	return nil
}
