package cmd

import (
	"ebmgo/bookfinder"
	"ebmgo/bookmanager"
	"ebmgo/config"
	"ebmgo/editor"
	"flag"
	"os"
)

func Import(call []string) error {
	flagSet := flag.NewFlagSet("import", flag.PanicOnError)
	skipEditFlag := flagSet.Bool("y", false, "Skip editing book metadata before import")
	helpFlag := flagSet.Bool("h", false, "Show help")

	flagSet.Parse(call)

	if *helpFlag {
		println("Usage: import [options] [directory]\n")
		println("Options:")
		flagSet.PrintDefaults()
		return nil
	}

	args := flagSet.Args()
	var path string
	if len(args) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		path = cwd
	} else {
		path = args[0]
	}

	return importBook(*skipEditFlag, path)

}

func importBook(skipEdit bool, path string) error {
	books, err := bookfinder.GetEbooks(path)
	if err != nil {
		return err
	}

	if !skipEdit {
		if err = editor.PrepareBooksForImport(books); err != nil {
			return err
		}
	}

	ebm, err := bookmanager.NewBookManager(bindPath(config.EBMGoLibraryDir))
	if err != nil {
		return err
	}
	defer ebm.Close()

	return ebm.ImportBooks(books)
}
