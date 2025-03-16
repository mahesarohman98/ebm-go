package main

import (
	"fmt"
	"os"
)

func cmdList() {
	println("EBM-GO\n")
	println("Simple Ebook Management Library In Go\n")
	println("Commands:")
	for name, command := range Apps {
		println("  ", name, " ", command.description)
	}
}

func main() {
	if len(os.Args) <= 1 {
		cmdList()
		return
	}

	cmdName := os.Args[1]
	cmd, ok := Apps[cmdName]
	if !ok {
		fmt.Print("Could not find apps \"" + cmdName + "\"")
	}
	args := os.Args

	e := cmd.run(args[2:])
	if e != nil {
		panic(e)
	}
}
