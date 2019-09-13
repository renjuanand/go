package main

import (
	"fmt"
	"os"
	"text/tabwriter"
)

type AboutCommand struct{}

func (c *AboutCommand) Execute(v *Vcli, args ...string) error {
	a := v.client.Client.ServiceContent.About
	tw := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "Name:\t%s\n", a.Name)
	fmt.Fprintf(tw, "Vendor:\t%s\n", a.Vendor)
	fmt.Fprintf(tw, "Version:\t%s\n", a.Version)
	fmt.Fprintf(tw, "Build:\t%s\n", a.Build)
	fmt.Fprintf(tw, "OS type:\t%s\n", a.OsType)
	fmt.Fprintf(tw, "API type:\t%s\n", a.ApiType)
	fmt.Fprintf(tw, "API version:\t%s\n", a.ApiVersion)
	fmt.Fprintf(tw, "Product ID:\t%s\n", a.ProductLineId)
	fmt.Fprintf(tw, "UUID:\t%s\n", a.InstanceUuid)
	tw.Flush()
	return nil
}
