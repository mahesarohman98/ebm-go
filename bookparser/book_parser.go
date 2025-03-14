package bookparser

import (
	"bytes"
	"errors"
	"io"
	"os"
)

var (
	ErrNotSupportMimeType = errors.New("unsupported mime type")
)

// File consist of file information such as filetype (pdf, epub) and path.
type File struct {
	Path string
	Type string
}

// Metadata consist of ebook metadata.
type Metadata struct {
	ISBN      string
	Title     string
	Authors   []string
	Publisher string
	Tags      []string
}

// BookParser is an instance of book info, consist of ebook metadata and file information.
type BookParser struct {
	File     File
	Metadata Metadata
}

// Parse parses ebook from path return BookInfo.
func Parse(path string) (BookParser, error) {
	f, err := os.Open(path)
	if err != nil {
		return BookParser{}, err
	}
	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return BookParser{}, err
	}

	reader := BookParser{File: File{Path: path}}

	if isPDF(f, fi.Size()) {
		reader.File.Type = "pdf"
		metadata, err := readMetadataFromPDF(path)
		if err != nil {
			return BookParser{}, err
		}

		reader.Metadata = metadata
	} else if isEpub(f, fi.Size()) {
		reader.File.Type = "epub"
		metadata, err := parseMetadataFromEpub(path)
		if err != nil {
			return BookParser{}, err
		}

		reader.Metadata = metadata
	} else if isMobi(f, fi.Size()) {
		reader.File.Type = "mobi"
		metadata, err := parseMetadataFromMobi(path)
		if err != nil {
			return BookParser{}, err
		}

		reader.Metadata = metadata
	} else {
		return BookParser{}, ErrNotSupportMimeType
	}

	return reader, nil
}

// isPDF check is mime type pdf.
func isPDF(f io.ReaderAt, size int64) bool {
	if size < 10 {
		return false // File too small to be a valid PDF
	}
	buf := make([]byte, 10)
	f.ReadAt(buf, 0)
	if !bytes.HasPrefix(buf, []byte("%PDF-")) || buf[7] < '0' || buf[7] > '7' || buf[8] != '\r' && buf[8] != '\n' {
		return false
	}
	return true
}

// isEpub check is mime type epub.
func isEpub(f io.ReaderAt, size int64) bool {
	if size < 4 {
		return false // File to small to be a valid Epub
	}
	buf := make([]byte, 4)
	f.ReadAt(buf, 0)
	return bytes.Equal(buf, []byte("PK\x03\x04"))
}

// isMobi check is mime type mobi.
func isMobi(f io.ReaderAt, size int64) bool {
	if size < 64 {
		return false
	}
	// Read 4 bytes at offset 60 (0x3C) where "BOOK" should be
	bookMarker := make([]byte, 4)
	_, err := f.ReadAt(bookMarker, 60)
	if err != nil {
		return false
	}
	// Read 4 bytes at offset 64 (0x40) where "MOBI" should be
	mobiMarker := make([]byte, 4)
	_, err = f.ReadAt(mobiMarker, 64)
	if err != nil {
		return false
	}

	// Check if the marker is "MOBI"
	return bytes.Equal(bookMarker, []byte("BOOK")) && bytes.Equal(mobiMarker, []byte("MOBI"))
}

func getTitleFromFilePath(filePath string) string {
	title := ""
	for i := len(filePath) - 1; i > 0; i-- {
		char := filePath[i]
		switch char {
		case '/':
			return title
		case '.':
			title = ""
		default:
			title = string(char) + title
		}
	}

	return title
}
