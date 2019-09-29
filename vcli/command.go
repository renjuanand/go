package main

import (
	"github.com/tatsushid/go-prettytable"
)

type Command interface {
	Execute(v *Vcli, args ...string) (*prettytable.Table, error)
}

// commands available for vCLI prompt
var Commands = map[string]Command{
	"about":   &AboutCommand{},
	"cl":      &ClCommand{},
	"dc":      &DcCommand{},
	"exit":    &ExitCommand{},
	"help":    &HelpCommand{},
	"hx":      &HxCommand{},
	"version": &VersionCommand{},
	"vm":      &VmCommand{},
	"quit":    &ExitCommand{},
}
