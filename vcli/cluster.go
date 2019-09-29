package main

import (
	_ "errors"
	"github.com/tatsushid/go-prettytable"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"path"
	_ "strings"
)

type ClCommand struct{}
type ClListCommand struct{}
type ClInfoCommand struct{}

const (
	CL_LIST = "list"
	CL_INFO = "info"
)

var clCommands = map[string]Command{
	CL_LIST: &ClListCommand{},
	CL_INFO: &ClInfoCommand{},
}

func (c *ClCommand) Execute(v *Vcli, args ...string) (*prettytable.Table, error) {
	if len(args) > 0 {
		cmd := args[0]
		options := args[1:]
		if fn, ok := clCommands[cmd]; ok {
			t, err := fn.Execute(v, options...)
			return t, err
		} else {
			Error("Unknown subcommand '%s' for vm\n", cmd)
		}
		return nil, nil
	}
	Message(c.Usage())
	return nil, nil
}

func (c *ClCommand) Usage() string {
	return `Usage: cl {command}

Commands:
	list
	info`
}

func (cmd *ClListCommand) Execute(cli *Vcli, args ...string) (*prettytable.Table, error) {
	clusters, err := GetClusterComputeResources(cli)
	if err != nil {
		return nil, err
	}

	tbl, err := prettytable.NewTable([]prettytable.Column{
		{Header: "No."},
		{Header: "Name", MinWidth: 6},
		{Header: "Path"},
		{Header: "Hosts"},
		{Header: "TotalCPU"},
		{Header: "TotalMemory"},
		{Header: "Cores"},
	}...)

	for index, cl := range clusters {
		// hostObjects, _ := cl.ComputeResource.Hosts(ctx)
		// hostSystems, _ := getHostSystems(cli, hostObjects)
		cr, _ := getComputeResource(cli, &cl.ComputeResource)
		summary := cr.Summary.GetComputeResourceSummary()
		tbl.AddRow(index+1, cl.Name(), cl.InventoryPath, summary.NumHosts, summary.TotalCpu, summary.TotalMemory, summary.NumCpuCores)
	}

	return tbl, nil
}

func (c *ClInfoCommand) Execute(cli *Vcli, args ...string) (*prettytable.Table, error) {
	Error("Not implemented")
	return nil, nil
}

func GetClusterComputeResources(cli *Vcli) ([]*object.ClusterComputeResource, error) {
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
		return nil, nil
	}

	objs := make(map[types.ManagedObjectReference]mo.Datacenter, len(datacenters))

	for _, o := range datacenters {
		objs[o.Reference()] = o
	}

	var clusters []*object.ClusterComputeResource

	for _, o := range objects {
		folders, err := o.Folders(ctx)

		if err != nil {
			return nil, err
		}

		finder.SetDatacenter(o)
		ccrs, _ := finder.ClusterComputeResourceList(ctx, path.Join(folders.HostFolder.InventoryPath, "*"))
		clusters = append(clusters, ccrs...)
	}
	return clusters, nil
}

func getComputeResource(cli *Vcli, cr *object.ComputeResource) (*mo.ComputeResource, error) {
	ctx := cli.ctx
	c := cli.client.Client
	props := []string{
		"summary",
	}

	var mocr mo.ComputeResource
	refs := make([]types.ManagedObjectReference, 0, 1)
	refs = append(refs, cr.Reference())
	pc := property.DefaultCollector(c)
	err := pc.Retrieve(ctx, refs, props, &mocr)

	if err != nil {
		return nil, err
	}

	return &mocr, nil
}

func getHostSystems(cli *Vcli, objects []*object.HostSystem) ([]mo.HostSystem, error) {
	ctx := cli.ctx
	c := cli.client.Client
	props := []string{
		"summary",
		"network",
	}

	refs := make([]types.ManagedObjectReference, 0, len(objects))
	for _, o := range objects {
		refs = append(refs, o.Reference())
	}

	var hosts []mo.HostSystem
	pc := property.DefaultCollector(c)
	err := pc.Retrieve(ctx, refs, props, &hosts)
	if err != nil {
		return nil, err
	}

	if len(hosts) <= 0 {
		return nil, nil
	}

	objs := make(map[types.ManagedObjectReference]mo.HostSystem, len(hosts))

	for _, o := range hosts {
		objs[o.Reference()] = o
	}

	return hosts, nil
}

func getNetworkProps(cli *Vcli, refs []types.ManagedObjectReference, props []string) ([]mo.Network, error) {
	ctx := cli.ctx
	c := cli.client.Client

	var network []mo.Network
	pc := property.DefaultCollector(c)
	err := pc.Retrieve(ctx, refs, props, &network)

	if err != nil {
		return nil, err
	}

	return network, nil
}

func GetMorProps(cli *Vcli, refs []types.ManagedObjectReference, props []string, dst interface{}) error {
	pc := property.DefaultCollector(cli.client.Client)
	if err := pc.Retrieve(cli.ctx, refs, props, dst); err != nil {
		return err
	}
	return nil
}
