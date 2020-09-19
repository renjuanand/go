package main

import (
	_ "fmt"
	prompt "github.com/c-bata/go-prompt"
	"strings"
)

var commands = []prompt.Suggest{
	{Text: "about", Description: "Display About info for HOST"},
	{Text: "cr", Description: "Cluster commands"},
	{Text: "dc", Description: "Datacenter commands"},
	{Text: "en", Description: "Extension commands"},
	{Text: "exit", Description: "Exit vcli"},
	{Text: "help", Description: "Show list of vcli commands"},
	{Text: "hx", Description: "HX commands"},
	{Text: "version", Description: "Show ESXi or vCenter version"},
	{Text: "vm", Description: "VM commands"},
	{Text: "quit", Description: "Exit vcli"},
}

func commandsCompleter(args []string) []prompt.Suggest {
	if len(args) <= 1 {
		return prompt.FilterHasPrefix(commands, args[0], true)
	}

	first := args[0]

	switch first {
	case "vm":
		second := args[1]
		if len(args) == 2 {
			subcommands := []prompt.Suggest{
				{Text: "destroy", Description: "Destroy VM"},
				{Text: "info", Description: "Show VM info"},
				{Text: "list", Description: "List all VMs"},
				{Text: "poweroff", Description: "Poweroff VM"},
				{Text: "poweron", Description: "Poweron VM"},
				{Text: "reset", Description: "Reset VM"},
			}
			return prompt.FilterHasPrefix(subcommands, second, true)
		}
	case "dc":
		second := args[1]
		if len(args) == 2 {
			subcommands := []prompt.Suggest{
				{Text: "list", Description: "List all datacenters"},
			}
			return prompt.FilterHasPrefix(subcommands, second, true)
		}
	case "en":
		second := args[1]
		if len(args) == 2 {
			subcommands := []prompt.Suggest{
				{Text: "list", Description: "List all extensions"},
				{Text: "info", Description: "Show details of an extension"},
				{Text: "register", Description: "Register an extension"},
				{Text: "unregister", Description: "Unregister extension(s)"},
			}
			return prompt.FilterHasPrefix(subcommands, second, true)
		}
	case "cr":
		second := args[1]
		if len(args) == 2 {
			subcommands := []prompt.Suggest{
				{Text: "list", Description: "List all clusters"},
				{Text: "info", Description: "Show info about a cluster"},
			}
			return prompt.FilterHasPrefix(subcommands, second, true)
		}
	case "hx":
		second := args[1]
		if len(args) == 2 {
			subcommands := []prompt.Suggest{
				{Text: "destroy", Description: "Destroy a given HX cluster"},
				{Text: "list", Description: "List all HX clusters"},
				{Text: "info", Description: "Show info about given HX cluster"},
				{Text: "summary", Description: "Show info about given HX cluster"},
			}
			return prompt.FilterHasPrefix(subcommands, second, true)
		}

	case "help":
		return []prompt.Suggest{}
	}
	return []prompt.Suggest{}
}

func completer(d prompt.Document) []prompt.Suggest {
	if d.TextBeforeCursor() == "" {
		return []prompt.Suggest{}
		// return commands
	}
	args := strings.Split(d.TextBeforeCursor(), " ")
	w := d.GetWordBeforeCursor()

	// If word before the cursor starts with "-", return options for subcommand
	if strings.HasPrefix(w, "-") {
		return optionCompleter(args, strings.HasPrefix(w, "--"))
	}

	return commandsCompleter(args)
}
