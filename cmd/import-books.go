package cmd

import (
	"context"
	"ebmgo/bookfinder"
	"ebmgo/bookmanager"
	"ebmgo/config"
	"ebmgo/editor"
	"flag"
	"os"
	"os/signal"
	"syscall"
)

func Import(call []string) error {
	flagSet := flag.NewFlagSet("import", flag.PanicOnError)
	skipEditFlag := flagSet.Bool("y", false, "Skip editing book metadata before import")
	recursiveFlag := flagSet.Bool("r", false, "import books recursively")
	workerFlag := flagSet.Int("w", 1, "set worker to import book. Default 1")
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

	return importBook(*workerFlag, *skipEditFlag, *recursiveFlag, path)

}

func importBook(worker int, skipEdit bool, recursive bool, path string) error {
	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Trap interrupt signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()

	books, err := bookfinder.GetEbooks(worker, recursive, path)
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

	err = ebm.ImportBooks(ctx, worker, books)
	return err
}
