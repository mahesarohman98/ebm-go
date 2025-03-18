package cmd

import (
	"ebmgo/bookmanager"
	"ebmgo/config"
	"flag"
	"fmt"
	"strconv"
	"strings"
)

func RemoveBooks(call []string) error {
	flagSet := flag.NewFlagSet("remove", flag.PanicOnError)
	idsFlag := flagSet.String("ids", "", "Book ID to remove. Separe by ','")
	helpFlag := flagSet.Bool("h", false, "Show help")

	flagSet.Parse(call)

	if *helpFlag {
		println("Usage: remove [options]\n")
		println("Options:")
		flagSet.PrintDefaults()
		return nil
	} else if *idsFlag == "" {
		return fmt.Errorf("ids is required")
	}

	var ids []int
	for _, sID := range strings.Split(*idsFlag, ",") {
		id, err := strconv.Atoi(sID)
		if err != nil {
			return fmt.Errorf("error parse flag ids: %v", err)
		}
		ids = append(ids, id)
	}

	return removeBooks(ids)
}

func removeBooks(ids []int) error {
	ebm, err := bookmanager.NewBookManager(bindPath(config.EBMGoLibraryDir))
	if err != nil {
		return err
	}
	defer ebm.Close()

	if err := ebm.RemoveBooks(ids); err != nil {
		return err
	}

	return nil
}
