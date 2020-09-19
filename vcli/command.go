package main

import (
	"github.com/tatsushid/go-prettytable"
)

type Command interface {
	Execute(v *Vcli, args ...string) (*prettytable.Table, error)
}

// commands available for vcli prompt
var Commands = map[string]Command{
	"about":   &AboutCommand{},
	"cr":      &CrCommand{},
	"dc":      &DcCommand{},
	"en":      &EnCommand{},
	"exit":    &ExitCommand{},
	"help":    &HelpCommand{},
	"hx":      &HxCommand{},
	"version": &VersionCommand{},
	"vm":      &VmCommand{},
	"quit":    &ExitCommand{},
}
