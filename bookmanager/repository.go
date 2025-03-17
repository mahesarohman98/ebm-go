package bookmanager

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type repository struct {
	db *sql.DB
}

func newRepository(db *sql.DB) *repository {
	return &repository{db: db}
}

func (repo *repository) CreateBooks(
	fn func() ([]Book, error),
) error {
	tx, err := repo.db.Begin()
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

	return repo.batchInsertBooks(tx, books)
}

func (repo *repository) batchInsertBooks(tx *sql.Tx, books []Book) error {
	now := time.Now()

	valueStrings := make([]string, 0)
	valueArgs := make([]interface{}, 0)
	param := 1
	for _, book := range books {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", param, param+1, param+2, param+3, param+4))
		valueArgs = append(valueArgs, nil)
		valueArgs = append(valueArgs, book.Title)
		valueArgs = append(valueArgs, book.ISBN)
		valueArgs = append(valueArgs, now)
		valueArgs = append(valueArgs, now)
		param += 5
	}

	if param <= 1 {
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
	startID := int(lastID) - len(books) + 1
	for i := range books {
		books[i].ID = startID + i
	}

	if err := repo.batchInsertFiles(tx, books); err != nil {
		return nil
	}

	if err := repo.batchInsertAuthors(tx, books); err != nil {
		return nil
	}

	if err := repo.batchInsertTags(tx, books); err != nil {
		return nil
	}

	return nil
}

func (repo *repository) batchInsertFiles(tx *sql.Tx, books []Book) error {
	now := time.Now()

	valuesString := make([]string, 0)
	valueArgs := make([]interface{}, 0)
	param := 1
	for i := range books {
		for _, file := range books[i].BookFiles {
			valuesString = append(valuesString, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", param, param+1, param+2, param+3, param+4))
			valueArgs = append(valueArgs, books[i].ID)
			valueArgs = append(valueArgs, file.FilePath)
			valueArgs = append(valueArgs, file.FileType)
			valueArgs = append(valueArgs, now)
			valueArgs = append(valueArgs, now)
			param += 5
		}
	}

	query := fmt.Sprintf(`
        INSERT INTO
            BookFiles (bookId, filePath, fileType, createDate, modifiedDate)
        VALUES %s
        `, strings.Join(valuesString, ","))
	_, err := tx.Exec(query, valueArgs...)

	return err
}

func (repo *repository) batchInsertAuthors(tx *sql.Tx, books []Book) error {
	authorValuesString := make([]string, 0)
	authorValueArgs := make([]interface{}, 0)
	authorParam := 1

	bookAuthorsString := make([]string, 0)
	bookAuthorsValueArgs := make([]interface{}, 0)
	bookAuthorsParam := 1
	for i := range books {
		for _, author := range books[i].Authors {
			if author == "" {
				continue
			}

			authorValuesString = append(authorValuesString, fmt.Sprintf("($%d)", authorParam))
			authorValueArgs = append(authorValueArgs, author)
			authorParam++

			bookAuthorsString = append(bookAuthorsString, fmt.Sprintf("($%d, $%d)", bookAuthorsParam, bookAuthorsParam+1))
			bookAuthorsValueArgs = append(bookAuthorsValueArgs, books[i].ID)
			bookAuthorsValueArgs = append(bookAuthorsValueArgs, author)
			bookAuthorsParam += 2
		}
	}

	if authorParam > 1 {
		query := fmt.Sprintf(`
            INSERT INTO Authors (author) VALUES %s ON CONFLICT(author) DO NOTHING
            `, strings.Join(authorValuesString, ","))
		if _, err := tx.Exec(query, authorValueArgs...); err != nil {
			return err
		}
	}

	if bookAuthorsParam > 1 {
		query := fmt.Sprintf(`
            INSERT INTO BookAuthors (bookId, author) 
                VALUES %s 
            ON CONFLICT(bookId, author) DO NOTHING
            `, strings.Join(bookAuthorsString, ","))
		if _, err := tx.Exec(query, bookAuthorsValueArgs...); err != nil {
			return err
		}
	}

	return nil
}

func (repo *repository) batchInsertTags(tx *sql.Tx, books []Book) error {
	bookTagsString := make([]string, 0)
	bookTagsValueArgs := make([]interface{}, 0)
	bookTagsParam := 1

	tagsValueStrings := make([]string, 0)
	tagsValueArgs := make([]interface{}, 0)
	tagsParam := 1

	for i := range books {
		for _, tag := range books[i].Tags {
			if tag == "" {
				continue
			}

			tagsValueStrings = append(tagsValueStrings, fmt.Sprintf("($%d)", tagsParam))
			tagsValueArgs = append(tagsValueArgs, tag)
			tagsParam++

			bookTagsString = append(bookTagsString, fmt.Sprintf("($%d, $%d)", bookTagsParam, bookTagsParam+1))
			bookTagsValueArgs = append(bookTagsValueArgs, books[i].ID)
			bookTagsValueArgs = append(bookTagsValueArgs, tag)
			bookTagsParam += 2
		}
	}

	if tagsParam > 1 {
		query := fmt.Sprintf(`
            INSERT INTO Tags (tag) VALUES %s ON CONFLICT(tag) DO NOTHING
            `, strings.Join(tagsValueStrings, ","))
		if _, err := tx.Exec(query, tagsValueArgs...); err != nil {
			return err
		}
	}

	if bookTagsParam > 1 {
		query := fmt.Sprintf(`
            INSERT INTO BookTags (bookId, tag) VALUES %s ON CONFLICT(bookId, tag) DO NOTHING
            `, strings.Join(bookTagsString, ","))
		if _, err := tx.Exec(query, bookTagsValueArgs...); err != nil {
			return err
		}
	}

	return nil

}

type bookDB struct {
	ID       int
	Title    string
	ISBN     string
	Author   string
	Tag      *string
	FilePath string
	FileType string
}

func (repo *repository) FindBooks(pattern string) ([]Book, error) {
	query := `
        SELECT
            b.bookId, b.title, b.isbn, 
            ba.author,
            bt.tag,
            bf.filePath, bf.fileType
        FROM Books b
            INNER JOIN BooksFts bfts ON bfts.bookId = b.bookId
			JOIN BookFiles bf USING(bookId)
			JOIN BookAuthors ba USING(bookId)
            LEFT JOIN  BookTags bt USING(bookId)
    `
	var args []interface{}
	if pattern != "" {
		query += "WHERE bfts.BooksFts MATCH $1"
		args = append(args, pattern)
	}

	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return []Book{}, fmt.Errorf("sql query error: %v", err)
	}

	books := []Book{}
	currentID := 0
	currentBook := Book{}
	for rows.Next() {
		b := bookDB{}
		if err := rows.Scan(&b.ID, &b.Title, &b.ISBN, &b.Author, &b.Tag, &b.FilePath, &b.FileType); err != nil {
			return []Book{}, err
		}

		if currentID == 0 {
			currentBook = NewBook(b.ISBN, b.Title, []string{}, "", []string{})
			currentBook.ID = b.ID
			currentID = b.ID
		} else if b.ID != currentID {
			books = append(books, currentBook)
			currentBook = NewBook(b.ISBN, b.Title, []string{}, "", []string{})
			currentBook.ID = b.ID
			currentID = b.ID
		}

		currentBook.AppendAuthors(b.Author)
		if b.Tag != nil {
			currentBook.AppendTag(*b.Tag)
		}
		currentBook.AppendFiles(b.FilePath, b.FileType)
	}
	books = append(books, currentBook)

	return books, nil
}

func rollback(tx *sql.Tx, books []Book) {
	tx.Rollback()
	for _, book := range books {
		for _, file := range book.BookFiles {
			os.Remove(file.FilePath)
		}
	}
}
