package main

import (
	"github.com/tatsushid/go-prettytable"
)

type HelpCommand struct{}

var HelpTopic = []struct {
	cmd         string
	description string
	example     string
}{

	{"about", "About info of ESXi or vCenter server", "about"},
	{"cr list", "Shows list of clusters", "cl list"},
	{"dc list", "Shows list of datacenters", "dc list"},
	{"hx list", "Shows list of HyperFlex clusters", "hx list"},
	{"hx info {cluster}", "Display about info for given HyperFlex cluster", "hx info Blr-Edge-CL"},
	{"hx destroy {cluster}", "Destroy a given HyperFlex cluster", "hx destroy Blr-Edge-CL"},
	{"version", "Version of ESXi or vCenter server", "version"},
	{"vm info {vm name}", "Display about info for given VM", "vm info ubuntu-vm1"},
	{"vm list", "Shows list of all virtual machines", "vm list"},
	{"vm destroy {vm1, vm2, ...}", "Destroy given list of virtual machines", "vm destroy vm1,vm2"},
	{"vm poweroff {vm1, vm2, ...}", "PowerOff given list of virtual machines", "vm poweroff vm1,vm2"},
	{"vm poweron {vm1, vm2, ...}", "PowerOn given list of virtual machines", "vm poweron vm1,vm2"},
	{"vm reset {vm1, vm2, ...}", "Reset given list of virtual machines", "vm reset vm1,vm2"},
	{"quit", "Quit vCLI", "quit"},
}

// 'help' command handler
func (c *HelpCommand) Execute(v *Vcli, args ...string) (*prettytable.Table, error) {
	tbl, err := prettytable.NewTable([]prettytable.Column{
		{Header: "", MinWidth: 20},
		{Header: "", MinWidth: 35},
		{Header: "", MinWidth: 25},
	}...)

	if err != nil {
		return nil, err
	}

	tbl.NoHeader = true
	tbl.AddRow(Key("Command"), Key("  Description"), Key("    Examples"))
	for _, ch := range HelpTopic {
		tbl.AddRow(HC(ch.cmd), HD(ch.description), HE(ch.example))
	}
	return tbl, nil
}
