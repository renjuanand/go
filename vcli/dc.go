package main

import (
	_ "context"
	"errors"
	"github.com/tatsushid/go-prettytable"
	"github.com/vmware/govmomi/find"
	_ "github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"path"
)

type DcCommand struct{}
type DcListCommand struct{}

const (
	DC_LIST = "list"
)

var dcCommands = map[string]Command{
	DC_LIST: &DcListCommand{},
}

func (c *DcCommand) Execute(v *Vcli, args ...string) (*prettytable.Table, error) {
	if len(args) > 0 {
		cmd := args[0]
		options := args[1:]
		if fn, ok := dcCommands[cmd]; ok {
			t, err := fn.Execute(v, options...)
			return t, err
		} else {
			Error("Unknown subcommand '%s' for vm\n", cmd)
		}
		return nil, nil
	}
	Usage(c.Usage())
	return nil, nil
}

func (cmd *DcCommand) Usage() string {
	return `Usage: dc [command]

Commands:
  list    List all datacenters`
}

func (cmd *DcListCommand) Execute(cli *Vcli, args ...string) (*prettytable.Table, error) {
	ctx := cli.ctx
	c := cli.client.Client

	finder := find.NewFinder(c, false)
	objects, err := finder.DatacenterList(ctx, "*")
	if err != nil {
		return nil, err
	}

	props := []string{
		"name",
		"vmFolder",
		"hostFolder",
		"datastoreFolder",
		"networkFolder",
		"datastore",
		"network",
	}

	refs := make([]types.ManagedObjectReference, 0, len(objects))
	for _, o := range objects {
		refs = append(refs, o.Reference())
	}

	var datacenters []mo.Datacenter
	pc := property.DefaultCollector(c)
	err = pc.Retrieve(ctx, refs, props, &datacenters)
	if err != nil {
		return nil, err
	}

	if len(datacenters) <= 0 {
		return nil, errors.New("No datacenters found")
	}

	objs := make(map[types.ManagedObjectReference]mo.Datacenter, len(datacenters))

	for _, o := range datacenters {
		objs[o.Reference()] = o
	}

	tbl, err := prettytable.NewTable([]prettytable.Column{
		{Header: "#"},
		{Header: "Name", MinWidth: 6},
		{Header: "Path"},
		{Header: "Hosts"},
		{Header: "Clusters"},
	}...)

	for i, o := range objects {
		dc := objs[o.Reference()]
		folders, err := o.Folders(ctx)

		if err != nil {
			return nil, err
		}

		finder.SetDatacenter(o)
		hosts, _ := finder.HostSystemList(ctx, path.Join(folders.HostFolder.InventoryPath, "*"))
		clusters, _ := finder.ClusterComputeResourceList(ctx, path.Join(folders.HostFolder.InventoryPath, "*"))
		tbl.AddRow(i+1, dc.Name, o.InventoryPath, len(hosts), len(clusters))
	}

	return tbl, nil
}
