package main

import (
	prompt "github.com/c-bata/go-prompt"
)

func optionCompleter(args []string, long bool) []prompt.Suggest {
	l := len(args)
	if l > 2 && args[1] == "list" {
		return optionHelp
	}

	return []prompt.Suggest{}
}

var optionHelp = []prompt.Suggest{
	{Text: "-grep"},
}
