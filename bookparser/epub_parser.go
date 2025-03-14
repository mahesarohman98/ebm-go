package bookparser

import "github.com/pirmd/epub"

func parseMetadataFromEpub(path string) (Metadata, error) {
	metadata, err := epub.GetMetadataFromFile(path)
	if err != nil {
		return Metadata{}, err
	}
	isbn := ""
	if len(metadata.Source) > 0 {
		isbn = metadata.Source[0]
	}
	title := ""
	if len(metadata.Title) > 0 {
		title = metadata.Title[0]
	}
	if title == "" {
		title = getTitleFromFilePath(path)
	}

	authors := []string{}
	if len(metadata.Creator) > 0 {
		for _, author := range metadata.Creator {
			name := author.FullName
			if name == "" {
				name = "Unknown"
			}
			authors = append(authors, name)
		}
	}
	publisher := ""
	if len(metadata.Publisher) > 0 {
		publisher = metadata.Publisher[0]
	}
	tags := []string{}
	for _, tag := range metadata.Subject {
		if tag != "" {
			tags = append(tags, tag)
		}
	}

	return Metadata{
		ISBN:      isbn,
		Title:     title,
		Authors:   authors,
		Publisher: publisher,
		Tags:      tags,
	}, nil
}
