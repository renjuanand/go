package main

import (
	"github.com/tatsushid/go-prettytable"
)

type VersionCommand struct{}

func (c *VersionCommand) Execute(v *Vcli, args ...string) (*prettytable.Table, error) {
	a := v.client.Client.ServiceContent.About
	tbl, err := prettytable.NewTable([]prettytable.Column{
		{Header: "Key"},
		{Header: "Value"},
	}...)

	if err != nil {
		return nil, err
	}

	tbl.NoHeader = true
	tbl.AddRow("Version:", a.Version)
	return tbl, nil
}
