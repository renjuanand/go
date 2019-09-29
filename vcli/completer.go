package main

import (
	_ "fmt"
	prompt "github.com/c-bata/go-prompt"
	"strings"
)

var commands = []prompt.Suggest{
	{Text: "about", Description: "Display About info for HOST."},
	{Text: "cl", Description: "Cluster commands."},
	{Text: "dc", Description: "Datacenter commands."},
	{Text: "exit", Description: "Exit this program"},
	{Text: "help", Description: "Show vCLI commands usage"},
	{Text: "version", Description: "Show ESXi or vCenter version."},
	{Text: "vm", Description: "VM commands."},
	{Text: "quit", Description: "Exit this program"},
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
				{Text: "list", Description: "List all VMs"},
				{Text: "info", Description: "Show VM info"},
				{Text: "poweroff", Description: "Poweroff VM"},
				{Text: "poweron", Description: "Poweron VM"},
				{Text: "destroy", Description: "Destroy VM"},
			}
			return prompt.FilterHasPrefix(subcommands, second, true)
		}
	case "dc":
		second := args[1]
		if len(args) == 2 {
			subcommands := []prompt.Suggest{
				{Text: "list", Description: "List all datacenters"},
				{Text: "info", Description: "Show info about a datacenter"},
			}
			return prompt.FilterHasPrefix(subcommands, second, true)
		}
	case "cl":
		second := args[1]
		if len(args) == 2 {
			subcommands := []prompt.Suggest{
				{Text: "list", Description: "List all clusters"},
				{Text: "info", Description: "Show info about a cluster"},
			}
			return prompt.FilterHasPrefix(subcommands, second, true)
		}

	case "help":
		return []prompt.Suggest{}
	}
	return []prompt.Suggest{}
}

func excludeOptions(args []string) ([]string, bool) {
	l := len(args)
	filtered := make([]string, 0, l)

	shouldSkipNext := []string{
		"-s",
		"--host",
		"--user",
		"-u",
		"--password",
		"-p",
	}

	var skipNextArg bool
	for i := 0; i < len(args); i++ {
		if skipNextArg {
			skipNextArg = false
			continue
		}

		for _, s := range shouldSkipNext {
			if strings.HasPrefix(args[i], s) {
				if strings.Contains(args[i], "=") {
					// we can specify option value like '-o=json'
					skipNextArg = false
				} else {
					skipNextArg = true
				}
				continue
			}
		}
		if strings.HasPrefix(args[i], "-") {
			continue
		}

		filtered = append(filtered, args[i])
	}
	return filtered, skipNextArg
}

func getPreviousOption(d prompt.Document) (cmd, option string, found bool) {
	args := strings.Split(d.TextBeforeCursor(), " ")
	l := len(args)
	if l >= 2 {
		option = args[l-2]
	}
	if strings.HasPrefix(option, "-") {
		return args[0], option, true
	}
	return "", "", false
}

func completeOptionArguments(d prompt.Document) ([]prompt.Suggest, bool) {
	//cmd, option, found := getPreviousOption(d)
	//_, _, found := getPreviousOption(d)
	//if !found {
	//	return []prompt.Suggest{}, false
	//}
	return []prompt.Suggest{}, false
}

func completer(d prompt.Document) []prompt.Suggest {
	if d.TextBeforeCursor() == "" {
		return []prompt.Suggest{}
	}
	args := strings.Split(d.TextBeforeCursor(), " ")
	w := d.GetWordBeforeCursor()

	// If word before the cursor starts with "-", returns CLI flag options.
	if strings.HasPrefix(w, "-") {
		return optionCompleter(args, strings.HasPrefix(w, "--"))
	}

	// Return suggestions for option
	if suggests, found := completeOptionArguments(d); found {
		return suggests
	}

	commandArgs, skipNext := excludeOptions(args)
	if skipNext {
		return []prompt.Suggest{}
	}

	return commandsCompleter(commandArgs)
}
