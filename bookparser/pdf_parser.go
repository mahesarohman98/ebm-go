package bookparser

import (
	"github.com/mahesarohman98/pdfinfo"
	"strings"
)

func readMetadataFromPDF(path string) (Metadata, error) {
	info, err := pdfinfo.Extract(path)
	if err != nil {
		return Metadata{}, err
	}

	title := info.Key("Title").Text()
	if title == "" {
		title = getTitleFromFilePath(path)
	}

	authors := strings.Split(info.Key("Author").Text(), "/")
	for i := range authors {
		if authors[i] == "" {
			authors[i] = "unknown"
		}
	}

	tags := []string{}
	for _, tag := range strings.Split(info.Key("Subject").Text(), "/") {
		if tag != "" {
			tags = append(tags, tag)
		}
	}

	return Metadata{
		ISBN:      info.Key("ISBN").Text(),
		Title:     title,
		Authors:   authors,
		Publisher: info.Key("Creator").Text(),
		Tags:      tags,
	}, nil
}
