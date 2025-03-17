package bookparser

import (
	mobipocket "github.com/clee/gobipocket"
)

func parseMetadataFromMobi(path string) (Metadata, error) {
	m, err := mobipocket.Open(path)
	if err != nil {
		return Metadata{Title: getTitleFromFilePath(path)}, nil
	}
	title := ""
	if len(m.Metadata["title"]) > 0 {
		title = m.Metadata["title"][0]
	}
	if title == "" {
		title = getTitleFromFilePath(path)
	}

	authors := m.Metadata["author"]
	for i := range authors {
		if authors[i] == "" {
			authors[i] = "Unknown"
		}
	}

	publisher := ""
	if len(m.Metadata["publisher"]) > 0 {
		publisher = m.Metadata["publisher"][0]
	}

	return Metadata{
		ISBN:      "",
		Title:     title,
		Authors:   authors,
		Publisher: publisher,
		Tags:      []string{},
	}, nil

}
