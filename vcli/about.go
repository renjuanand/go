package main

import (
	"github.com/tatsushid/go-prettytable"
	_ "text/tabwriter"
)

type AboutCommand struct{}

// 'about' command handler
func (c *AboutCommand) Execute(v *Vcli, args ...string) (*prettytable.Table, error) {
	a := v.client.Client.ServiceContent.About
	tbl, err := prettytable.NewTable([]prettytable.Column{
		{Header: "Key"},
		{Header: "Value"},
	}...)

	if err != nil {
		return nil, err
	}

	aboutTbl := []KeyValue{
		{"Name", a.Name},
		{"Vendor", a.Vendor},
		{"Version:", a.Version},
		{"Build:", a.Build},
		{"OS type:", a.OsType},
		{"API type:", a.ApiType},
		{"API version:", a.ApiVersion},
		{"Product ID:", a.ProductLineId},
		{"UUID:", a.InstanceUuid},
	}

	tbl.NoHeader = true
	for _, k := range aboutTbl {
		tbl.AddRow(k.key, k.value)
	}
	return tbl, nil
}
