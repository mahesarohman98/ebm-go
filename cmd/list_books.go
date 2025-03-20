package cmd

import (
	"ebmgo/bookmanager"
	"ebmgo/config"
	"flag"
	"fmt"
	"os"
	"strings"
)

func ListBooks(call []string) error {
	flagSet := flag.NewFlagSet("list", flag.PanicOnError)
	queryFlag := flagSet.String("s", "", "Filter the results by the search query")
	formatFlag := flagSet.String("f", "title,authors", "The fields to display when listing books in the db. Available fields: title, authors, formats. Default: title,authors.")
	helpFlag := flagSet.Bool("h", false, "Show help")

	flagSet.Parse(call)

	if *helpFlag {
		println("Usage: list [options]\n")
		println("Options:")
		flagSet.PrintDefaults()
		return nil
	}

	return listBooks(*queryFlag, *formatFlag)
}

func isValidFormat(value string) bool {
	switch value {
	case "title":
		return true
	case "authors":
		return true
	case "formats":
		return true
	default:
		return false
	}
}

func parseFormat(formatFlag string) ([]string, error) {
	var format []string
	var value []rune
	appendFormat := func() error {
		v := string(value)
		if !isValidFormat(v) {
			return fmt.Errorf("unknown formats: %s", v)
		}

		format = append(format, v)
		value = []rune{}

		return nil
	}

	for _, r := range formatFlag {
		switch r {
		case ',':
			if err := appendFormat(); err != nil {
				return []string{}, err
			}
		case ' ':
			continue
		default:
			value = append(value, r)
		}
	}
	if len(value) > 0 {
		if err := appendFormat(); err != nil {
			return []string{}, err
		}
	}

	return format, nil
}

func listBooks(query string, formatFlag string) error {
	ebm, err := bookmanager.NewBookManager(bindPath(config.EBMGoLibraryDir))
	if err != nil {
		return err
	}
	defer ebm.Close()

	books, err := ebm.GetBooks(query)
	if err != nil {
		return err
	}

	format, err := parseFormat(formatFlag)
	if err != nil {
		return err
	}

	// Print title from given format
	fmt.Fprintf(os.Stdout, "%-5s", "ID")
	for _, f := range format {
		switch f {
		case "title":
			fmt.Fprintf(os.Stdout, "%-75s", "Title")
		case "authors":
			fmt.Fprintf(os.Stdout, "%-60s", "Author(s)")
		case "formats":
			fmt.Fprintf(os.Stdout, "%-60s", "Formats")
		}
	}
	fmt.Fprintln(os.Stdout, "")

	// Print data from given format
	for _, b := range books {
		fmt.Fprintf(os.Stdout, "%-5d", b.ID)
		for _, f := range format {
			switch f {
			case "title":
				fmt.Fprintf(os.Stdout, "%-75s", b.Title)
			case "authors":
				fmt.Fprintf(os.Stdout, "%-60s", strings.Join(b.Authors, " & "))
			case "formats":
				fmt.Fprintf(os.Stdout, "%v-60s", b.BookFiles)
			}
		}
		fmt.Fprintln(os.Stdout, "")
	}

	return nil
}
