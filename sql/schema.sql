CREATE TABLE IF NOT EXISTS Books(
    bookID INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL DEFAULT 'Unknown' COLLATE NOCASE,
    isbn TEXT DEFAULT '' COLLATE NOCASE,
    createDate TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedDate TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS BookFiles(
    bookId INTEGER NOT NULL,
    filePath TEXT,
    fileType TEXT,
    createDate TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedDate TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(bookId) REFERENCES Books(bookId),
    UNIQUE(bookId, filePath, fileType)
);

CREATE TABLE IF NOT EXISTS Authors(
    author TEXT NOT NULL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS BookAuthors(
    bookId INTEGER NOT NULL,
    author TEXT NOT NULL,
    PRIMARY KEY(bookId, author),
    FOREIGN KEY (bookId) REFERENCES Books(bookId),
    FOREIGN KEY (author) REFERENCES Authors(author)
);

CREATE TABLE IF NOT EXISTS Tags(
    tag TEXT NOT NULL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS BookTags(
    bookId INTEGER NOT NULL,
    tag TEXT NOT NULL,
    PRIMARY KEY(bookId, tag),
    FOREIGN KEY (bookId) REFERENCES Books(bookId),
    FOREIGN KEY (tag) REFERENCES Tags(tag)
);

CREATE VIRTUAL TABLE IF NOT EXISTS BooksFts USING fts5(
   bookId, title, isbn, createDate, modifiedDate 
);

CREATE TRIGGER InsertBookFts
    AFTER INSERT ON Books
BEGIN
    INSERT INTO BooksFts(bookId, title, isbn, createDate,modifiedDate)
VALUES (NEW.bookId, New.title, NEW.isbn, NEW.createDate, NEW.modifiedDate);
END;

CREATE TRIGGER UpdateBookFts
    AFTER UPDATE ON Books
BEGIN
    UPDATE BooksFts
    SET
        title = NEW.title,
        isbn = NEW.isbn,
        modifiedDate = NEW.modifiedDate
    WHERE
        bookId = NEW.bookId;
END;

CREATE TRIGGER DeleteBookFts
    AFTER DELETE ON Books
BEGIN
    DELETE FROM BooksFts
    WHERE bookId = OLD.bookId;
END;
