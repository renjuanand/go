package main

import (
	"github.com/tatsushid/go-prettytable"
)

type HelpCommand struct{}

func (c *HelpCommand) Execute(v *Vcli, args ...string) (*prettytable.Table, error) {
	Message("vcli commands help")
	return nil, nil
}
