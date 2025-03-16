package main

import "ebmgo/cmd"

var Apps map[string]run = map[string]run{
	"import": {description: "import books from given path", run: cmd.Import},
	"list":   {description: "list books in ebm directory", run: cmd.ListBooks},
}

type run struct {
	description string
	run         func(call []string) error
}
