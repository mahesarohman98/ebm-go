package main

import "ebmgo/cmd"

var Apps map[string]Command = map[string]Command{
	"import": cmd.Import,
}

type Command func(call []string) error
