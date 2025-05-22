package bookmanager

import (
	"context"
	"database/sql"
	"fmt"
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
	ctx context.Context,
	fn func() ([]*Book, error),
	rollbackFn func(),
) error {
	tx, err := repo.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	var books []*Book
	defer func() {
		if err != nil || ctx.Err() != nil {
			tx.Rollback()
			rollbackFn()
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

	return repo.batchInsertBooks(ctx, tx, books)
}

func (repo *repository) batchInsertBooks(ctx context.Context, tx *sql.Tx, books []*Book) error {
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

	res, err := tx.ExecContext(ctx, query, valueArgs...)
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

	if err := repo.batchInsertFiles(ctx, tx, books); err != nil {
		return nil
	}

	if err := repo.batchInsertAuthors(ctx, tx, books); err != nil {
		return nil
	}

	if err := repo.batchInsertTags(ctx, tx, books); err != nil {
		return nil
	}

	return nil
}

func (repo *repository) batchInsertFiles(ctx context.Context, tx *sql.Tx, books []*Book) error {
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
	_, err := tx.ExecContext(ctx, query, valueArgs...)

	return err
}

func (repo *repository) batchInsertAuthors(ctx context.Context, tx *sql.Tx, books []*Book) error {
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
		if _, err := tx.ExecContext(ctx, query, authorValueArgs...); err != nil {
			return err
		}
	}

	if bookAuthorsParam > 1 {
		query := fmt.Sprintf(`
            INSERT INTO BookAuthors (bookId, author) 
                VALUES %s 
            ON CONFLICT(bookId, author) DO NOTHING
            `, strings.Join(bookAuthorsString, ","))
		if _, err := tx.ExecContext(ctx, query, bookAuthorsValueArgs...); err != nil {
			return err
		}
	}

	return nil
}

func (repo *repository) batchInsertTags(ctx context.Context, tx *sql.Tx, books []*Book) error {
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
		if _, err := tx.ExecContext(ctx, query, tagsValueArgs...); err != nil {
			return err
		}
	}

	if bookTagsParam > 1 {
		query := fmt.Sprintf(`
            INSERT INTO BookTags (bookId, tag) VALUES %s ON CONFLICT(bookId, tag) DO NOTHING
            `, strings.Join(bookTagsString, ","))
		if _, err := tx.ExecContext(ctx, query, bookTagsValueArgs...); err != nil {
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

func parseBooks(bookDBs []bookDB) []Book {
	books := []Book{}
	currentID := 0
	currentBook := Book{}
	for _, b := range bookDBs {
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

	return books
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
		return []Book{}, fmt.Errorf("query FindBooks error: %v", err)
	}
	var booksDBs []bookDB
	for rows.Next() {
		b := bookDB{}
		if err := rows.Scan(&b.ID, &b.Title, &b.ISBN, &b.Author, &b.Tag, &b.FilePath, &b.FileType); err != nil {
			return []Book{}, err
		}
		booksDBs = append(booksDBs, b)
	}
	books := parseBooks(booksDBs)

	return books, nil
}

func (repo *repository) getBooks(ids []int) ([]Book, error) {
	placeholders := ""
	var arg []interface{}
	for i, id := range ids {
		if i == len(ids)-1 {
			placeholders += fmt.Sprintf("$%d", i+1)
		} else {
			placeholders += fmt.Sprintf("$%d, ", i+1)
		}
		arg = append(arg, id)
	}

	query := fmt.Sprintf(`
        SELECT 
            b.bookId, b.title, b.isbn,
            ba.author,
            bt.tag,
            bf.filePath, bf.fileType
        FROM Books b
			JOIN BookFiles bf USING(bookId)
			JOIN BookAuthors ba USING(bookId)
            LEFT JOIN  BookTags bt USING(bookId)
        WHERE
            b.bookId IN (%s)

        `, placeholders)

	rows, err := repo.db.Query(query, arg...)
	if err != nil {
		return []Book{}, fmt.Errorf("query getBooks error: %v", err)
	}

	var booksDBs []bookDB
	for rows.Next() {
		b := bookDB{}
		if err := rows.Scan(&b.ID, &b.Title, &b.ISBN, &b.Author, &b.Tag, &b.FilePath, &b.FileType); err != nil {
			return []Book{}, err
		}
		booksDBs = append(booksDBs, b)
	}
	books := parseBooks(booksDBs)

	return books, nil
}

func (repo *repository) RemoveBooks(
	ids []int,
	fn func(b []Book) error,
	rollbackFn func(),
) error {
	_, err := repo.db.Exec("BEGIN IMMEDIATE")
	if err != nil {
		return fmt.Errorf("start tx begin immediate error: %v", err)
	}

	var books []Book
	defer func() {
		if err != nil {
			repo.db.Exec("ROLLBACK")
			rollbackFn()
		} else {
			repo.db.Exec("COMMIT")
		}
	}()

	books, err = repo.getBooks(ids)
	if err != nil {
		return err
	}
	err = fn(books)
	if err != nil {
		return err
	}

	placeholders := ""
	var arg []interface{}
	for i, id := range ids {
		if i == len(ids)-1 {
			placeholders += fmt.Sprintf("$%d", i+1)
		} else {
			placeholders += fmt.Sprintf("$%d, ", i+1)
		}
		arg = append(arg, id)
	}
	query := fmt.Sprintf("DELETE FROM books WHERE bookId IN ( %s )", placeholders)

	_, err = repo.db.Exec(query, arg...)
	if err != nil {
		return err
	}

	return nil
}
