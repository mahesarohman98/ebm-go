package cmd

import (
	"ebmgo/bookmanager"
	"ebmgo/config"
	"flag"
	"fmt"
	"os"
)

func ListBooks(call []string) error {
	flagSet := flag.NewFlagSet("list", flag.PanicOnError)
	queryFlag := flagSet.String("s", "", "Filter the results by the search query")
	helpFlag := flagSet.Bool("h", false, "Show help")

	flagSet.Parse(call)

	if *helpFlag {
		println("Usage: findbooks [options]\n")
		println("Options:")
		flagSet.PrintDefaults()
		return nil
	}

	return listBooks(*queryFlag)
}

func listBooks(query string) error {
	ebm, err := bookmanager.NewBookManager(bindPath(config.EBMGoLibraryDir))
	if err != nil {
		return err
	}
	defer ebm.Close()

	books, err := ebm.GetBooks(query)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, "id\t", "title")
	for _, b := range books {
		fmt.Fprintln(os.Stdout, b.ID, "\t", b.Title)
	}
	return nil
}
