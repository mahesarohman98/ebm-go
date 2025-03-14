package cmd

import (
	"ebmgo/bookfinder"
	"ebmgo/bookmanager"
	"ebmgo/config"
	"ebmgo/editor"
	"flag"
	"os"
)

var (
	flagSet  = flag.NewFlagSet("import", flag.PanicOnError)
	helpFlag = flagSet.Bool("h", false, "Show help")
)

func Import(call []string) error {
	flagSet.Parse(call)

	if *helpFlag {
		help()
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

	return importBook(path)

}

func importBook(path string) error {
	files, err := bookfinder.GetEbooks(path)
	if err != nil {
		return err
	}

	books, err := editor.PrepareBooksForImport(files)
	if err != nil {
		return err
	}

	ebm, err := bookmanager.NewBookManager(bindPath(config.EBMGoLibraryDir))
	if err != nil {
		return err
	}

	return ebm.ImportBooks(books)
}

func help() {
	println("Usage: import [options] [directory]\n")
	println("Options:")
	flagSet.PrintDefaults()
}
