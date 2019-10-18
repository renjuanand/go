package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/tatsushid/go-prettytable"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type VmCommand struct{}
type VmDestroyCommand struct{}
type VmInfoCommand struct{}
type VmListCommand struct{}
type VmPowerOffCommand struct{}
type VmPowerOnCommand struct{}
type VmResetCommand struct{}

const (
	VM_DESTROY  = "destroy"
	VM_INFO     = "info"
	VM_LIST     = "list"
	VM_POWEROFF = "poweroff"
	VM_POWERON  = "poweron"
	VM_RESET    = "reset"
)

var vmCommands = map[string]Command{
	VM_DESTROY:  &VmDestroyCommand{},
	VM_INFO:     &VmInfoCommand{},
	VM_LIST:     &VmListCommand{},
	VM_POWEROFF: &VmPowerOffCommand{},
	VM_POWERON:  &VmPowerOnCommand{},
	VM_RESET:    &VmResetCommand{},
}

// type vmActionFunc func(string, context.Context) (*mo.Task, error)

type vmAction struct {
	action             string
	startActionMessage string
}

var vmActions = map[string]vmAction{
	VM_DESTROY:  vmAction{"Destroy", "Destroying"},
	VM_POWEROFF: vmAction{"PowerOff", "Powering off"},
	VM_POWERON:  vmAction{"PowerOn", "Powering on"},
	VM_RESET:    vmAction{"Reset", "Resetting"},
}

func (c *VmCommand) Execute(v *Vcli, args ...string) (*prettytable.Table, error) {
	if len(args) > 0 {
		cmd := args[0]
		options := args[1:]
		if fn, ok := vmCommands[cmd]; ok {
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

func (c *VmCommand) Usage() string {
	return `Usage: vm {command}

Commands:
	destroy
	info
	list
	poweroff
	poweron
	reset`
}

func (cmd *VmListCommand) Execute(cli *Vcli, args ...string) (*prettytable.Table, error) {
	ctx := cli.ctx
	c := cli.client.Client
	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return nil, err
	}

	var option, filter string
	enableFilter := false

	if len(args) > 0 {
		option = args[0]
		if option == "--grep" && len(args) == 2 {
			filter = args[1]
			enableFilter = true
		}
	}

	// Retrieve summary property for all machines
	var vms []mo.VirtualMachine
	props := []string{"summary"}
	props = append(props, "guest.ipAddress")
	err = v.Retrieve(ctx, []string{"VirtualMachine"}, props, &vms)
	if err != nil {
		return nil, err
	}

	tbl, err := prettytable.NewTable([]prettytable.Column{
		{Header: "No"},
		{Header: "Name"},
		{Header: "IP Address"},
		{Header: "State"},
	}...)

	// tbl.Separator = " | "
	for index, vm := range vms {
		var ip string
		if vm.Guest != nil {
			ip = vm.Guest.IpAddress
		}

		if enableFilter {
			r, _ := regexp.Compile(filter)
			//if strings.Contains(vm.Summary.Config.Name, filter) {
			if r.MatchString(vm.Summary.Config.Name) {
				tbl.AddRow(index+1, vm.Summary.Config.Name, ip, string(vm.Summary.Runtime.PowerState))
			}
		} else {
			tbl.AddRow(index+1, vm.Summary.Config.Name, ip, string(vm.Summary.Runtime.PowerState))
		}
	}

	return tbl, nil
}

func (c *VmInfoCommand) Usage() string {
	return `Usage: vm info {vm name}`
}

func (cmd *VmInfoCommand) Execute(cli *Vcli, args ...string) (*prettytable.Table, error) {
	if len(args) <= 0 {
		return nil, errors.New(cmd.Usage())
	}

	vmName := args[0]
	ctx := cli.ctx
	c := cli.client.Client
	m := view.NewManager(c)

	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return nil, err
	}

	var vms []mo.VirtualMachine
	err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary"}, &vms)

	if err != nil {
		return nil, err
	}

	var targetVm *mo.VirtualMachine
	for index, vm := range vms {
		if vm.Summary.Config.Name == vmName || strconv.Itoa(index+1) == vmName {
			targetVm = &vm
			break
		}
	}

	if targetVm == nil {
		return nil, errors.New("Virtual machine '" + vmName + "' is not found")
	}

	// vmRef := object.NewVirtualMachine(c, targetVm.Reference())

	tbl, err := prettytable.NewTable([]prettytable.Column{
		{Header: "Key"},
		{Header: "Value"},
	}...)

	if err != nil {
		return nil, err
	}

	s := targetVm.Summary

	infoTbl := []KeyValue{
		{"Name", s.Config.Name},
		{"UUID:", s.Config.Uuid},
		{"Guest name:", s.Config.GuestFullName},
		{"Memory:", strconv.FormatInt(int64(s.Config.MemorySizeMB), 10) + "MB"},
		{"CPU:", strconv.FormatInt(int64(s.Config.NumCpu), 10) + " vCPU(s)"},
		{"Power state:", fmt.Sprintf("%s", s.Runtime.PowerState)},
		{"Boot time:", fmt.Sprintf("%s", s.Runtime.BootTime)},
		{"IP address:", s.Guest.IpAddress},
	}

	tbl.NoHeader = true
	for _, k := range infoTbl {
		tbl.AddRow(k.key, k.value)
	}

	return tbl, nil
}

func (c *VmPowerOnCommand) Usage() string {
	return `Usage: vm poweron vm-name1, vm-name2, ... OR
       vm poweron 1,2,...`
}

func (c *VmPowerOnCommand) Execute(cli *Vcli, args ...string) (*prettytable.Table, error) {
	if len(strings.Join(args, "")) == 0 {
		return nil, errors.New(c.Usage())
	}
	err := executeVmCommand(VM_POWERON, cli, args...)
	return nil, err
}

func (c *VmPowerOffCommand) Usage() string {
	return `Usage: vm poweroff vm-name1, vm-name2, ... OR
       vm poweroff 1,2,...`
}

func (c *VmPowerOffCommand) Execute(cli *Vcli, args ...string) (*prettytable.Table, error) {
	if len(strings.Join(args, "")) == 0 {
		return nil, errors.New(c.Usage())
	}
	err := executeVmCommand(VM_POWEROFF, cli, args...)
	return nil, err
}

func (c *VmDestroyCommand) Usage() string {
	return `Usage: vm destroy vm-name1, vm-name2, ... OR
       vm destroy 1,2,...`
}

func (c *VmDestroyCommand) Execute(cli *Vcli, args ...string) (*prettytable.Table, error) {
	if len(strings.Join(args, "")) == 0 {
		return nil, errors.New(c.Usage())
	}
	err := executeVmCommand(VM_DESTROY, cli, args...)
	return nil, err
}

func (c *VmResetCommand) Usage() string {
	return `Usage: vm reset vm-name1, vm-name2, ... OR
       vm reset 1,2,...`
}

func (c *VmResetCommand) Execute(cli *Vcli, args ...string) (*prettytable.Table, error) {
	if len(strings.Join(args, "")) == 0 {
		return nil, errors.New(c.Usage())
	}
	err := executeVmCommand(VM_RESET, cli, args...)
	return nil, err
}

func executeVmCommand(action string, cli *Vcli, args ...string) error {
	var vmArgs []string
	if len(args) > 0 {
		vmArgs = strings.Split(args[0], ",")
	} else {
		return errors.New("usage error")
	}

	ctx := cli.ctx
	c := cli.client.Client
	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return err
	}

	var vms, actionableVms []mo.VirtualMachine
	// Retrieve summary property for all machines
	props := []string{"summary"}
	err = v.Retrieve(ctx, []string{"VirtualMachine"}, props, &vms)
	if err != nil {
		return err
	}

	for _, name := range vmArgs {
		for index, vm := range vms {
			if vm.Summary.Config.Name == name || strconv.Itoa(index+1) == name {
				actionableVms = append(actionableVms, vm)
			}
		}
	}

	if len(actionableVms) == 0 {
		return errors.New(fmt.Sprintf("%v not found", vmArgs))
	}

	Spinner.Stop()
	var wg sync.WaitGroup
	channel := make(chan string)

	go func() {
		for m := range channel {
			Infoln(m)
		}
	}()

	for _, vm := range actionableVms {
		wg.Add(1)
		go func(vm interface{}) error {
			var machine mo.VirtualMachine = vm.(mo.VirtualMachine)

			defer wg.Done()
			vmRef := object.NewVirtualMachine(c, machine.Reference())
			vmName := machine.Summary.Config.Name
			channel <- fmt.Sprintf("%s '%s'...", vmActions[action].startActionMessage, vmName)
			err = doVmAction(vmRef, action, ctx)

			if err != nil {
				channel <- fmt.Sprintf("%s", err.Error())
				return err
			}

			if err == nil {
				channel <- fmt.Sprintf("%s completed for '%s'", vmActions[action].action, vmName)
			} else {
				channel <- fmt.Sprintf("Failed to %s vm '%s': %s", vmActions[action].action, vmName, err.Error())
			}
			return nil
		}(vm)

	}

	wg.Wait()
	close(channel)
	// Wait for 1 sec to show prompt properly, after showing up all messages from channel
	time.Sleep(1 * time.Second)
	return nil
}

func doVmAction(v *object.VirtualMachine, action string, ctx context.Context) error {
	var task *object.Task
	var err error

	switch action {
	case VM_POWERON:
		task, err = v.PowerOn(ctx)
	case VM_POWEROFF:
		task, err = v.PowerOff(ctx)
	case VM_DESTROY:
		task, err = v.Destroy(ctx)
	case VM_RESET:
		task, err = v.Reset(ctx)
	}

	if err != nil {
		return err
	}

	_, err = task.WaitForResult(ctx, nil)
	if err != nil {
		return err
	}

	if action == VM_POWERON {
		_, err = v.WaitForIP(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
