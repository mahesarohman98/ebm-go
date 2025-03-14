package main

import (
	"fmt"
	"os"
)

func cmdList() {
	println("List of commands:\n")
	for name := range Apps {
		print(name, ", ")
	}
	println("")
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

	e := cmd(args[2:])
	if e != nil {
		panic(e)
	}
}
