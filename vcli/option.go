package main

import (
	prompt "github.com/c-bata/go-prompt"
)

func optionCompleter(args []string, long bool) []prompt.Suggest {
	l := len(args)
	if l > 2 && ((args[0] == "vm" && args[1] == "list") || (args[0] == "hx" && args[1] == "info")) {
		return optionHelp
	}

	return []prompt.Suggest{}
}

var optionHelp = []prompt.Suggest{
	{Text: "-grep"},
}
