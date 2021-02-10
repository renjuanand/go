package main

import (
	_ "context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/tatsushid/go-prettytable"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
	"strings"
)

type EnCommand struct{}
type EnListCommand struct{}
type EnInfoCommand struct{}
type EnRegisterCommand struct{}
type EnUnregisterCommand struct{}

const (
	EN_LIST             = "list"
	EN_INFO             = "info"
	EN_REGISTER         = "register"
	EN_UNREGISTER       = "unregister"
	MAX_DESCRIPTION_LEN = 40
)

var enCommands = map[string]Command{
	EN_LIST:       &EnListCommand{},
	EN_INFO:       &EnInfoCommand{},
	EN_REGISTER:   &EnRegisterCommand{},
	EN_UNREGISTER: &EnUnregisterCommand{},
}

func (c *EnCommand) Execute(v *Vcli, args ...string) (*prettytable.Table, error) {
	if len(args) > 0 {
		cmd := args[0]
		options := args[1:]
		if fn, ok := enCommands[cmd]; ok {
			t, err := fn.Execute(v, options...)
			return t, err
		} else {
			Error("Unknown subcommand '%s' for en\n", cmd)
		}
		return nil, nil
	}
	Usage(c.Usage())
	return nil, nil
}

func (c *EnCommand) Usage() string {
	return `Usage: en [command]

Commands:
  list         List all extensions
  info         Show details of an extension
  register     Register extension
  unregister   Unregister extension
`
}

func (cmd *EnListCommand) Execute(cli *Vcli, args ...string) (*prettytable.Table, error) {
	ctx := cli.ctx
	c := cli.client.Client

	m, err := object.GetExtensionManager(c)
	if err != nil {
		return nil, err
	}

	list, err := m.List(ctx)
	if err != nil {
		return nil, err
	}

	exts := make(map[string]types.Extension)
	for _, e := range list {
		exts[e.Key] = e
	}

	var filter string
	listCmd := flag.NewFlagSet("list", flag.ContinueOnError)
	listGrep := listCmd.String("grep", "", "Search pattern")
	listCmd.Parse(args)

	if listGrep != nil {
		filter = *listGrep
	}

	tbl, err := prettytable.NewTable([]prettytable.Column{
		{Header: "#"},
		{Header: "Name"},
		{Header: "Version"},
		{Header: "Description"},
		{Header: "Company"},
	}...)

	for index, e := range list {
		desc := e.Description.GetDescription().Summary
		if len(desc) > MAX_DESCRIPTION_LEN {
			desc = desc[:MAX_DESCRIPTION_LEN] + "..."
		}
		/*
			if e.Key == "com.vmware.ovf" {
				serverInfo := e.Server
				for _, s := range serverInfo {
					fmt.Println(s.Url)
					fmt.Println(s.ServerThumbprint)
				}
			}
		*/
		if listGrep != nil && !strings.Contains(e.Key, filter) {
			continue
		}
		//tbl.AddRow(index+1, e.Key, e.Version, e.Description.GetDescription().Summary, e.Type, e.Company)
		tbl.AddRow(index+1, e.Key, e.Version, desc, e.Company)
	}

	return tbl, nil
}

func (c *EnInfoCommand) Usage() string {
	return `Usage: en info <extension-key>

Examples:
  en info com.vmware.ovf
`
}

func (cmd *EnInfoCommand) Execute(cli *Vcli, args ...string) (*prettytable.Table, error) {
	if len(args) <= 0 {
		Usage(cmd.Usage())
		return nil, nil
	}

	key := args[0]
	ctx := cli.ctx
	c := cli.client.Client

	m, err := object.GetExtensionManager(c)
	if err != nil {
		return nil, err
	}

	list, err := m.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, e := range list {
		if e.Key == key {
			data, err := json.MarshalIndent(&e, "", "  ")
			if err != nil {
				return nil, err
			}
			Spinner.Stop()
			fmt.Printf("%s\n", data)
			return nil, nil
		}
	}

	return nil, errors.New("Extension '" + key + "' is not found")
}

func (cmd *EnRegisterCommand) Execute(cli *Vcli, args ...string) (*prettytable.Table, error) {
	return nil, nil
}

func (cmd *EnUnregisterCommand) Usage() string {
	return `Usage: en unregister <extension-key>

Examples:
  en unregister com.vmware.ovf
`
}

func (cmd *EnUnregisterCommand) Execute(cli *Vcli, args ...string) (*prettytable.Table, error) {
	if len(args) <= 0 {
		Usage(cmd.Usage())
		return nil, nil
	}

	key := args[0]
	ctx := cli.ctx
	c := cli.client.Client

	m, err := object.GetExtensionManager(c)
	if err != nil {
		return nil, err
	}

	if err = m.Unregister(ctx, key); err != nil {
		return nil, err
	}

	fmt.Printf("'%s' has been successfully unregistered\n", key)
	return nil, nil
}
