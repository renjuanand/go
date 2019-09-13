package main

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"
)

type VersionCommand struct{}

func (c *VersionCommand) Execute(v *Vcli, args ...string) error {
	if v == nil {
		return errors.New("Failed to execute version command")
	}

	a := v.client.Client.ServiceContent.About
	tw := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "Version:\t%s\n", a.Version)
	tw.Flush()
	return nil
}
