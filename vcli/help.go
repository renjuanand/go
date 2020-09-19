package main

import (
	"github.com/tatsushid/go-prettytable"
)

type HelpCommand struct{}

// 'help' command handler
func (c *HelpCommand) Execute(v *Vcli, args ...string) (*prettytable.Table, error) {
	if Spinner.Active() {
		Spinner.Stop()
	}

	tbl, err := prettytable.NewTable([]prettytable.Column{
		{Header: "", MinWidth: 20},
		{Header: "", MinWidth: 35},
		{Header: "", MinWidth: 25},
	}...)

	if err != nil {
		return nil, err
	}

	tbl.NoHeader = true
	tbl.AddRow("Command", "Description", "Example(s)")
	tbl.AddRow("------------------------------",
		"---------------------------------------------------------------------", "-----------------------")
	tbl.AddRow("about", "About info of ESXi or vCenter host", "about")
	tbl.AddRow("cr list", "Shows list of clusters", "cr list")
	tbl.AddRow("dc list", "Shows list of datacenters", "dc list")
	tbl.AddRow("en list [-grep string]", "List all extensions", "en list")
	tbl.AddRow("", "Use -grep option to filter extensions by key", "en list -grep vmware")
	tbl.AddRow("hx list", "Shows list of HX clusters", "hx list")
	tbl.AddRow("hx info [-grep string] NAME", "Display about info of HX clusters", "hx info all")
	tbl.AddRow("", "NAME can be 'all' OR cluster names or numbers separated by comma", "hx info BLR-EDGE")
	tbl.AddRow("", "Use -grep option to filter cluster info by", "hx info BLR-EDGE,HX-CL2")
	tbl.AddRow("", "Name, Build, Version, Model, Serial No. and CIP", "hx info 1,2")
	tbl.AddRow("", "", "hx info -grep 4.0(1a) all")
	tbl.AddRow("", "", "hx info -grep UCSB-B200-M5 all")
	tbl.AddRow("", "", "hx info -grep FCH2206V1NG all")
	tbl.AddRow("hx destroy NAME", "Destroy a given HX cluster", "hx destroy BLR-EDGE")
	tbl.AddRow("version", "Shows ESXi or vCenter version", "version")
	tbl.AddRow("vm list [-grep string]", "Shows list of all virtual machines", "vm list")
	tbl.AddRow("", "Use -grep option to filter vm list by VM name, IP Address and Folder", "vm list -grep 10.64.55.177")
	tbl.AddRow("", "", "vm list -grep install-upgrade-ui")
	tbl.AddRow("vm info NAME", "Display about info of given virtual machine", "vm info ubuntu-vm1")
	tbl.AddRow("vm destroy NAME1[,NAME2, ...]", "Destroy virtual machines", "vm destroy Win2K16")
	tbl.AddRow("vm poweroff NAME1[,NAME2, ...]", "Power off virtual machines", "vm poweroff Win2K16")
	tbl.AddRow("vm poweron NAME1[,NAME2, ...]", "Power on virtual machines", "vm poweron LinuxVM")
	tbl.AddRow("vm reset NAME1[,NAME2, ...]", "Reset virtual machines", "vm reset Ubuntu18.04")
	tbl.AddRow("quit", "Quit vcli", "quit")

	return tbl, nil
}
