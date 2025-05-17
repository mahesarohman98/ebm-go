package cmd

import (
	"ebmgo/bookfinder"
	"ebmgo/bookmanager"
	"ebmgo/config"
	"ebmgo/editor"
	"flag"
	"log"
	"os"
	"time"
)

func Import(call []string) error {
	flagSet := flag.NewFlagSet("import", flag.PanicOnError)
	skipEditFlag := flagSet.Bool("y", false, "Skip editing book metadata before import")
	recursiveFlag := flagSet.Bool("r", false, "import books recursively")
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

	return importBook(*skipEditFlag, *recursiveFlag, path)

}

func importBook(skipEdit bool, recursive bool, path string) error {
	now := time.Now()
	books, err := bookfinder.GetEbooks(recursive, path)
	if err != nil {
		return err
	}
	log.Println("GetEbooks took", time.Since(now))

	if !skipEdit {
		now := time.Now()
		if err = editor.PrepareBooksForImport(books); err != nil {
			return err
		}
		log.Println("PrepareBooksForImport took", time.Since(now))
	}

	ebm, err := bookmanager.NewBookManager(bindPath(config.EBMGoLibraryDir))
	if err != nil {
		return err
	}
	defer ebm.Close()

	now = time.Now()
	err = ebm.ImportBooks(books)
	log.Println("PrepareBooksForImport took", time.Since(now))
	return err
}
