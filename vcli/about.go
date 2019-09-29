package main

import (
	"github.com/tatsushid/go-prettytable"
	_ "text/tabwriter"
)

type AboutCommand struct{}

func (c *AboutCommand) Execute(v *Vcli, args ...string) (*prettytable.Table, error) {
	a := v.client.Client.ServiceContent.About
	tbl, err := prettytable.NewTable([]prettytable.Column{
		{Header: "Key"},
		{Header: "Value"},
	}...)

	if err != nil {
		return nil, err
	}

	tbl.NoHeader = true
	tbl.AddRow("Name:", a.Name)
	tbl.AddRow("Vendor:", a.Vendor)
	tbl.AddRow("Version:", a.Version)
	tbl.AddRow("Build:", a.Build)
	tbl.AddRow("OS type:", a.OsType)
	tbl.AddRow("API type:", a.ApiType)
	tbl.AddRow("API version:", a.ApiVersion)
	tbl.AddRow("Product ID:", a.ProductLineId)
	tbl.AddRow("UUID:", a.InstanceUuid)
	return tbl, nil
}
