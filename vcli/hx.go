package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/tatsushid/go-prettytable"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"io/ioutil"
	"net/http"
	_ "regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type HxCommand struct{}
type HxListCommand struct{}
type HxInfoCommand struct{}
type HxDestroyCommand struct{}

const (
	HX_LIST    = "list"
	HX_INFO    = "info"
	HX_SUMMARY = "summary"
	HX_DESTROY = "destroy"

	CLIENT_TIMEOUT         = 20
	SC_MGMT_NETWORK        = "Storage Controller Management Network"
	ERR_NO_SC_MGMT_NETWORK = "No Storage Controller Management Network found"
)

var hxCommands = map[string]Command{
	HX_LIST:    &HxListCommand{},
	HX_INFO:    &HxInfoCommand{},
	HX_SUMMARY: &HxInfoCommand{},
	HX_DESTROY: &HxDestroyCommand{},
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	UserId       int
	Username     string
}

type ClusterAbout struct {
	Uuid                 string
	Hypervisor           string
	ApiVersion           string
	ProductVersion       string
	ModelNumber          string
	SerialNumber         string
	Name                 string
	FullName             string
	Build                string
	DisplayVersion       string
	EncryptionSupported  bool
	ReplicationSupported bool
	UpgradeSupported     string
}

type ClusterDetail struct {
	Name                  string
	DataReplicationFactor string
	ClusterAccessPolicy   string
	NumNodesConfigured    int
	NumNodesOnline        int
	ClusterIpAddress      string
	ClusterType           string
	ZoneType              string
	AllFlash              bool
	EncryptionEnabled     bool
	ReplicationEnabled    bool
}

type ClusterHealth struct {
	Uuid                      string
	State                     string
	DataReplicationCompliance string
}

type ClusterStats struct {
	SpaceStatus                string
	RawCapacityInBytes         int
	TotalCapacityInBytes       int
	UsedCapacityInBytes        int
	FreeCapacityInBytes        int
	CompressionSavings         float64
	DeduplicationSavings       float64
	TotalSavings               float64
	EnospaceState              string
	BytesToResumeIO            int
	BytesReclaimable           int
	BytesToFreeToClearEnospace int
}

type ClusterTime struct {
	CreateTime     int64
	UptimeInSecs   int64
	DowntimeInSecs int64
}

type ClusterSummary struct {
	Overview struct {
		Config struct {
			Name   string
			MgmtIp struct {
				Addr string
			}
		}
		About  ClusterAbout
		Detail ClusterDetail
		Stats  ClusterStats
		Health ClusterHealth
		Time   ClusterTime
	}
}

type NetworkAddress struct {
	Fqdn string `json:"fqdn"`
	Ip   string `json:"ip"`
}

type ClusterNetwork struct {
	ClusterMgmtIpAddress NetworkAddress `json:"clusterMgmtIpAddress"`
	ClusterDataIpAddress NetworkAddress `json:"clusterDataIpAddress"`
	WitnessNode          NetworkAddress `json:"witnessNode"`
}

func (c *HxCommand) Execute(v *Vcli, args ...string) (*prettytable.Table, error) {
	if len(args) > 0 {
		cmd := args[0]
		options := args[1:]
		if fn, ok := hxCommands[cmd]; ok {
			t, err := fn.Execute(v, options...)
			return t, err
		} else {
			Error("Unknown subcommand '%s' for hx\n", cmd)
		}
		return nil, nil
	}
	Usage(c.Usage())
	return nil, nil
}

func (c *HxCommand) Usage() string {
	return `Usage: hx [command]

Commands:
  list       List all HX clusters registered in this VC
  info       Display cluster summary of HX cluster(s)
  summary    Display cluster summary of HX cluster(s)
  destroy    Destroy a HX cluster
`
}

func (cmd *HxListCommand) Execute(cli *Vcli, args ...string) (*prettytable.Table, error) {
	ctx := cli.ctx
	c := cli.client.Client
	pc := property.DefaultCollector(c)

	clusters, err := GetClusterComputeResources(cli)
	if err != nil {
		return nil, err
	}

	if len(clusters) <= 0 {
		return nil, errors.New("No clusters found")
	}

	var hxClusters []*object.ClusterComputeResource

	for _, clr := range clusters {
		hostObjects, _ := clr.ComputeResource.Hosts(ctx)
		refs := make([]types.ManagedObjectReference, 0, len(hostObjects))
		for _, o := range hostObjects {
			refs = append(refs, o.Reference())
		}

		var hostSystems []mo.HostSystem
		err = pc.Retrieve(ctx, refs, []string{"summary", "network"}, &hostSystems)
		if err != nil {
			continue
		}

		ctlvmFound := false

		for _, host := range hostSystems {
			var networks []mo.Network
			err := pc.Retrieve(ctx, host.Network, []string{"vm"}, &networks)
			if err != nil {
				continue
			}
			for _, nw := range networks {
				var vms []mo.VirtualMachine
				err := pc.Retrieve(ctx, nw.Vm, []string{"name"}, &vms)
				if err != nil {
					continue
				}
				for _, vm := range vms {
					//fmt.Println(vm.Name)
					if strings.HasPrefix(vm.Name, "stCtlVM") {
						ctlvmFound = true
						break
					}
				}

				if ctlvmFound {
					break
				}
			}

			if ctlvmFound {
				break
			}
		}

		if ctlvmFound {
			hxClusters = append(hxClusters, clr)
		}
	}

	if len(hxClusters) <= 0 {
		return nil, errors.New("No HX clusters found")
	}

	tbl, err := prettytable.NewTable([]prettytable.Column{
		{Header: "#"},
		{Header: "Name", MinWidth: 6},
		{Header: "Path"},
		{Header: "Hosts"},
		{Header: "TotalCPU"},
		{Header: "Cores"},
		{Header: "TotalMemory"},
	}...)

	for index, cl := range hxClusters {
		cr, _ := GetComputeResource(cli, &cl.ComputeResource)
		summary := cr.Summary.GetComputeResourceSummary()
		tbl.AddRow(index+1, cl.Name(), cl.InventoryPath, summary.NumHosts, getCpuInGHz(summary.TotalCpu), summary.NumCpuCores, getMemoryInGB(summary.TotalMemory))
	}

	return tbl, nil
}

func (cmd *HxInfoCommand) Usage() string {
	return `Usage: hx info [options] all OR clusternames OR numbers separated by comma

Display summary info of HX cluster(s) registered in this VC

Options:
  -grep=pattern   Filter cluster summary info based on given search pattern
                  Available filter fields are Name, Version, Build, SerialNumber, ModelNumber and CIP

Examples:
  hx info all
  hx info 1,2,3
  hx info hx-blr-cl,edge-cl
  hx info -grep 3.5.2g all
  hx info -grep=240C all
  hx info -grep=UCSB-B200-M5 all
	`
}

func (cmd *HxInfoCommand) Execute(cli *Vcli, args ...string) (*prettytable.Table, error) {
	infoCmd := flag.NewFlagSet("info", flag.ContinueOnError)
	infoGrep := infoCmd.String("grep", "", "Search pattern")
	infoCmd.Parse(args)

	if len(infoCmd.Args()) == 0 {
		Usage(cmd.Usage())
		return nil, nil
	}

	clusterName := strings.Join(infoCmd.Args(), "")
	clusters, err := GetClusterComputeResources(cli)
	if err != nil {
		return nil, err
	}

	if len(clusters) <= 0 {
		return nil, errors.New("No clusters found")
	}

	var targetClusters []*object.ClusterComputeResource
	var hsl []*ClusterSummary
	if clusterName == "all" {
		targetClusters = clusters
	} else {
		clusterNames := strings.Split(strings.Trim(clusterName, " "), ",")
		for _, cname := range clusterNames {
			cname = strings.Trim(cname, " ")
			found := false
			for index, clr := range clusters {
				if clr.Name() == cname || strconv.Itoa(index+1) == cname {
					targetClusters = append(targetClusters, clr)
					found = true
					break
				}
			}
			if !found {
				Errorln("cluster '" + cname + "' doesn't exist")
			}
		}
	}

	if len(targetClusters) == 0 {
		return nil, errors.New("No clusters found")
	}

	var wg sync.WaitGroup
	for _, clr := range targetClusters {
		wg.Add(1)
		go func(clr interface{}) {
			c := clr.(*object.ClusterComputeResource)
			defer wg.Done()
			hs, err := getClusterInfo(cli, c)
			if err == nil && hs != nil {
				hsl = append(hsl, hs)
			} else if err.Error() != ERR_NO_SC_MGMT_NETWORK {
				Errorln("["+c.Name()+"]:", err)
			}
		}(clr)
	}
	wg.Wait()

	var filteredSummaryList []*ClusterSummary
	if infoGrep != nil {
		// r, _ := regexp.Compile(*infoGrep)
		pattern := *infoGrep
		for _, hx := range hsl {
			if strings.Contains(hx.Overview.Detail.Name, pattern) || strings.Contains(hx.Overview.About.DisplayVersion, pattern) ||
				strings.Contains(hx.Overview.About.Build, pattern) || strings.Contains(hx.Overview.Config.MgmtIp.Addr, pattern) ||
				strings.Contains(hx.Overview.About.SerialNumber, pattern) || strings.Contains(hx.Overview.About.ModelNumber, pattern) {
				filteredSummaryList = append(filteredSummaryList, hx)
			}
		}
	} else {
		filteredSummaryList = hsl
	}

	tbl, err := prettytable.NewTable([]prettytable.Column{
		{Header: "Key", MinWidth: 12},
		{Header: "Value"},
	}...)

	if err != nil {
		return nil, err
	}

	tbl.NoHeader = true
	for index, hx := range filteredSummaryList {
		if hx.Overview.Detail.Name == "" {
			continue
		}
		tbl.AddRow("Name:", hx.Overview.Detail.Name)
		tbl.AddRow("Version:", hx.Overview.About.DisplayVersion)
		tbl.AddRow("Build:", hx.Overview.About.Build)
		tbl.AddRow("CIP:", hx.Overview.Config.MgmtIp.Addr)
		tbl.AddRow("State:", hx.Overview.Health.State)
		tbl.AddRow("UUID:", hx.Overview.About.Uuid)
		tbl.AddRow("AllFlash:", hx.Overview.Detail.AllFlash)
		tbl.AddRow("SerialNumber:", hx.Overview.About.SerialNumber)
		tbl.AddRow("ModelNumber:", hx.Overview.About.ModelNumber)
		tbl.AddRow("AccessPolicy:", hx.Overview.Detail.ClusterAccessPolicy)
		tbl.AddRow("ReplicationFactor:", hx.Overview.Detail.DataReplicationFactor)
		tbl.AddRow("Uptime:", getUptimeString(hx.Overview.Time.UptimeInSecs))
		tbl.AddRow("Total Capacity:", getStorageCapacityInTB(hx.Overview.Stats.TotalCapacityInBytes))
		tbl.AddRow("Available Capacity:", getStorageCapacityInTB(hx.Overview.Stats.FreeCapacityInBytes))
		if len(filteredSummaryList) > 1 && (index+1) != len(filteredSummaryList) {
			tbl.AddRow("-------------------", "-------------------------------------")
		}
	}
	return tbl, nil
}

func getControllerIp(cli *Vcli, hxCluster *object.ClusterComputeResource) (string, error) {
	ctx := cli.ctx
	c := cli.client.Client
	pc := property.DefaultCollector(c)

	var mocr mo.ComputeResource
	crRef := hxCluster.ComputeResource.Reference()
	err := pc.RetrieveOne(ctx, crRef, []string{"network"}, &mocr)
	if err != nil {
		return "", err
	}

	var networks []mo.Network
	err = pc.Retrieve(ctx, mocr.Network, []string{"name", "vm"}, &networks)
	if err != nil {
		return "", err
	}

	for _, network := range networks {
		if network.Name != SC_MGMT_NETWORK {
			continue
		}
		var vms []mo.VirtualMachine
		err := pc.Retrieve(ctx, network.Vm, []string{"name", "guest", "resourcePool"}, &vms)
		if err != nil {
			return "", err
		}

		for _, vm := range vms {
			if strings.HasPrefix(vm.Name, "stCtlVM") && vm.Guest != nil && vm.Guest.Net != nil {
				var rp mo.ResourcePool
				err := pc.RetrieveOne(ctx, *vm.ResourcePool, []string{"owner"}, &rp)
				if err != nil {
					return "", err
				}

				if crRef.Value != rp.Owner.Value {
					continue
				}
				for _, nic := range vm.Guest.Net {
					if nic.Connected && nic.Network == SC_MGMT_NETWORK {
						if len(nic.IpAddress) > 1 {
							ctrlIp := nic.IpAddress[0]
							return ctrlIp, nil
						}
					}
				}
			}
		}
	}

	return "", errors.New(ERR_NO_SC_MGMT_NETWORK)
}

func getClusterInfo(cli *Vcli, hxCluster *object.ClusterComputeResource) (*ClusterSummary, error) {
	ctrlIp, err := getControllerIp(cli, hxCluster)
	if err != nil {
		return nil, err
	}

	hs, err := getHxSummary(cli.auth, ctrlIp)
	if err != nil {
		return nil, err
	}

	return hs, nil
}

func (cmd *HxDestroyCommand) Usage() string {
	return `Usage: hx destroy cluster-name

Destroys a HX cluster

Examples:
  hx destroy 3Node-cluster
`
}

func (cmd *HxDestroyCommand) Execute(cli *Vcli, args ...string) (*prettytable.Table, error) {
	if len(args) == 0 {
		Usage(cmd.Usage())
		return nil, nil
	}

	clusterName := args[0]
	ctx := cli.ctx
	c := cli.client.Client
	pc := property.DefaultCollector(c)

	clusters, err := GetClusterComputeResources(cli)
	if err != nil {
		return nil, err
	}

	if len(clusters) == 0 {
		return nil, errors.New("No clusters found")
	}

	var hxCluster *object.ClusterComputeResource
	for _, clr := range clusters {
		// if clr.Name() == clusterName || strconv.Itoa(index+1) == clusterName {
		if clr.Name() == clusterName {
			hxCluster = clr
			break
		}
	}

	if hxCluster == nil {
		return nil, errors.New("'" + clusterName + "' doesn't exist")
	}

	hostObjects, err := hxCluster.ComputeResource.Hosts(ctx)
	if err != nil {
		return nil, err
	}

	dsObjects, err := hxCluster.ComputeResource.Datastores(ctx)
	if err != nil {
		return nil, err
	}

	dsMap := make(map[types.ManagedObjectReference]*object.Datastore, len(hostObjects))
	for _, o := range dsObjects {
		dsMap[o.Reference()] = o
	}

	refs := make([]types.ManagedObjectReference, 0, len(hostObjects))
	hsMap := make(map[types.ManagedObjectReference]*object.HostSystem, len(hostObjects))

	for _, ho := range hostObjects {
		hsMap[ho.Reference()] = ho
		refs = append(refs, ho.Reference())
	}

	var hostSystems []mo.HostSystem
	err = pc.Retrieve(ctx, refs, []string{"summary", "datastore", "network", "vm"}, &hostSystems)
	if err != nil {
		return nil, err
	}

	// Remove Controller VMs from each host
	if len(hostSystems) > 0 {
		removeControllerVms(cli, &hostSystems[0])
	}

	for _, host := range hostSystems {
		hostName := host.Summary.Config.Name
		hns, err := hsMap[host.Reference()].ConfigManager().NetworkSystem(ctx)

		if err != nil {
			Errorln(err)
			continue
		}

		var mns mo.HostNetworkSystem
		err = pc.RetrieveOne(ctx, hns.Reference(), []string{"networkInfo.vnic", "networkInfo.vswitch", "networkInfo.portgroup"}, &mns)
		if err != nil {
			Errorln(err)
			continue
		}

		Spinner.Prefix = ""
		// Remove Virtual Nics
		removeVirtualNics(ctx, hns, mns.NetworkInfo)

		// Remove PortGroups
		removePortGroups(ctx, hns, mns.NetworkInfo)

		// Remove Virtual Switches
		removeVirtualSwitches(ctx, hns, mns.NetworkInfo)

		// Remove springpath datastore from host
		if host.Datastore != nil {
			var datastores []mo.Datastore
			err = pc.Retrieve(ctx, host.Datastore, []string{"name"}, &datastores)
			if err != nil {
				Errorln("Failed to find datastores for host '" + hostName + "' : " + err.Error())
			} else {
				for _, ds := range datastores {
					if strings.HasPrefix(ds.Name, "SpringpathDS") {
						hostObj := hsMap[host.Reference()]
						hds, err := hostObj.ConfigManager().DatastoreSystem(ctx)
						if err != nil {
							Errorln("Failed to find datastore system: ", err.Error())
						} else {
							err = hds.Remove(ctx, dsMap[ds.Reference()])
							if err != nil {
								Errorln("Failed to delete datastore '" + ds.Name + "' :" + err.Error())
							}
						}
						break
					}
				}
			}
		}
	}

	// Don't remove datacenter, multiple clusters can come under
	// a datacenter
	//err = removeDatacenter(cli, hxCluster)

	// Remove cluster
	task, err := hxCluster.Destroy(ctx)
	if err != nil {
		Errorln("Failed to destroy cluster '" + hxCluster.Name() + "' : " + err.Error())
	} else {
		_, err := task.WaitForResult(ctx, nil)
		if err != nil {
			Errorln("Failed to destroy cluster '" + hxCluster.Name() + "' : " + err.Error())
		}
	}
	return nil, err
}

func removeDatacenter(cli *Vcli, cr *object.ClusterComputeResource) error {
	ctx := cli.ctx
	c := cli.client.Client
	pc := property.DefaultCollector(c)
	var mcr mo.ComputeResource
	var mf mo.Folder
	var mdc mo.Datacenter

	err := pc.RetrieveOne(ctx, cr.ComputeResource.Reference(), []string{"parent"}, &mcr)
	if err != nil {
		return err
	}

	err = pc.RetrieveOne(ctx, *mcr.Entity().Parent, []string{"parent"}, &mf)
	if err != nil {
		return err
	}

	err = pc.RetrieveOne(ctx, *mf.Entity().Parent, []string{"name"}, &mdc)
	if err != nil {
		return err
	}

	dcObj := *object.NewDatacenter(c, mdc.Reference())
	task, err := dcObj.Destroy(ctx)
	_, err = task.WaitForResult(ctx, nil)
	return err
}

func removeControllerVms(cli *Vcli, host *mo.HostSystem) {
	ctx := cli.ctx
	c := cli.client.Client
	pc := property.DefaultCollector(c)
	var ctrlVms []mo.VirtualMachine
	var networks []mo.Network

	err := pc.Retrieve(ctx, host.Network, []string{"name", "vm"}, &networks)
	if err != nil {
		Errorln("Failed to find networks: " + err.Error())
	} else {
		for _, nw := range networks {
			if nw.Name == "Storage Controller Management Network" {
				var vms []mo.VirtualMachine
				err := pc.Retrieve(ctx, nw.Vm, []string{"name"}, &vms)
				if err != nil {
					continue
				}
				for _, vm := range vms {
					if strings.HasPrefix(vm.Name, "stCtlVM") {
						ctrlVms = append(ctrlVms, vm)
					}
				}

				if len(ctrlVms) > 0 {
					break
				}
			}
		}
	}

	var wg sync.WaitGroup

	for _, vm := range ctrlVms {
		wg.Add(1)
		go func(vm interface{}) {
			var machine mo.VirtualMachine = vm.(mo.VirtualMachine)
			defer wg.Done()
			vmRef := object.NewVirtualMachine(c, machine.Reference())
			err := doVmAction(vmRef, VM_POWEROFF, ctx)
			if err != nil {
				Errorln("Failed to poweroff vm '" + machine.Name + "' : " + err.Error())
			} else {
				err := vmRef.Unregister(ctx)
				if err != nil {
					Errorln("Failed to unregister vm '" + machine.Name + "' : " + err.Error())
				}
			}
		}(vm)
	}
	wg.Wait()
}

func removeVirtualNics(ctx context.Context, hns *object.HostNetworkSystem, ni *types.HostNetworkInfo) {
	if ctx == nil || hns == nil || ni == nil {
		return
	}
	for _, nic := range ni.Vnic {
		if nic.Portgroup == "Storage Hypervisor Data Network" {
			err := hns.RemoveVirtualNic(ctx, nic.Device)
			if err != nil {
				Errorln("Failed to remove: '" + nic.Device + "' vNic: " + err.Error())
			}
		}
	}

}

func removePortGroups(ctx context.Context, hns *object.HostNetworkSystem, ni *types.HostNetworkInfo) {
	if ctx == nil || hns == nil || ni == nil {
		return
	}
	for _, pg := range ni.Portgroup {
		if pg.Spec.Name == "Storage Controller Data Network" ||
			pg.Spec.Name == "Storage Controller Replication Network" ||
			pg.Spec.Name == "Storage Controller Management Network" ||
			pg.Spec.Name == "Storage Hypervisor Data Network" {
			err := hns.RemovePortGroup(ctx, pg.Spec.Name)
			if err != nil {
				Errorln("Failed to remove: '" + pg.Spec.Name + "' portgroup: " + err.Error())
			}
		}
	}
}

func removeVirtualSwitches(ctx context.Context, hns *object.HostNetworkSystem, ni *types.HostNetworkInfo) {
	if ctx == nil || hns == nil || ni == nil {
		return
	}
	for _, s := range ni.Vswitch {
		if s.Name == "vmotion" || s.Name == "vswitch-hx-vm-network" || s.Name == "vswitch-hx-storage-data" {
			err := hns.RemoveVirtualSwitch(ctx, s.Name)
			if err != nil {
				Errorln("Failed to remove: '" + s.Name + "' vswitch: " + err.Error())
			}
		}
	}
}

type HxRequest struct {
	host   string
	auth   *AuthResponse
	client *http.Client
}

func (r *HxRequest) Login(c *Credentials) error {
	reqBody, err := json.Marshal(map[string]string{
		"username":      c.username,
		"password":      c.password,
		"client_id":     "HxGuiClient",
		"client_secret": "Sunnyvale",
		"redirect_uri":  "hx",
	})

	respBody, err := r.Post("/aaa/v1/auth?grant_type=password", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	authResponse := AuthResponse{}
	if err = json.Unmarshal(respBody, &authResponse); err != nil {
		return err
	}

	r.auth = &authResponse
	return nil
}

func (r *HxRequest) Logout() error {
	if r.auth == nil {
		return errors.New("Not logged in")
	}
	reqBody, err := json.Marshal(map[string]string{
		"access_token":  r.auth.AccessToken,
		"refresh_token": r.auth.RefreshToken,
		"token_type":    r.auth.TokenType,
	})

	_, err = r.Post("/aaa/v1/revoke", bytes.NewBuffer(reqBody))
	return err
}

func (r *HxRequest) Get(api string) ([]byte, error) {
	return r.request("GET", api, bytes.NewBuffer([]byte{}))
}

func (r *HxRequest) Post(api string, reqBody *bytes.Buffer) ([]byte, error) {
	return r.request("POST", api, reqBody)
}

func (r *HxRequest) request(method string, api string, reqBody *bytes.Buffer) ([]byte, error) {
	protocol := "https://"
	url := protocol + r.host + api
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Agent", "HXConnect")
	req.Header.Set("Accept-Language", "en")

	if r.auth != nil && r.auth.TokenType != "" && r.auth.AccessToken != "" {
		req.Header.Set("Authorization", r.auth.TokenType+" "+r.auth.AccessToken)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

func getHxSummary(auth *Credentials, host string) (*ClusterSummary, error) {
	timeout := time.Duration(CLIENT_TIMEOUT * time.Second)
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	hxReq := &HxRequest{
		host:   host,
		client: client,
	}

	err := hxReq.Login(auth)
	if err != nil {
		return nil, err
	}

	aboutResponse, err := hxReq.Get("/coreapi/v1/clusters/1/about")
	if err != nil {
		return nil, err
	}

	detailResponse, err := hxReq.Get("/coreapi/v1/clusters/1/detail")
	if err != nil {
		return nil, err
	}

	networkResponse, err := hxReq.Get("/coreapi/v1/clusters/1/network")
	if err != nil {
		return nil, err
	}

	timeResponse, err := hxReq.Get("/coreapi/v1/clusters/1/time")
	if err != nil {
		return nil, err
	}

	statsResponse, err := hxReq.Get("/coreapi/v1/clusters/1/stats")
	if err != nil {
		return nil, err
	}

	healthResponse, err := hxReq.Get("/coreapi/v1/clusters/1/health")
	if err != nil {
		return nil, err
	}

	// We're done with all requests, logout hx session
	hxReq.Logout()

	summary := ClusterSummary{}
	clusterAbout := ClusterAbout{}
	if err = json.Unmarshal(aboutResponse, &clusterAbout); err != nil {
		Errorln(err.Error())
	} else {
		summary.Overview.About = clusterAbout
	}

	clusterDetail := ClusterDetail{}
	if err = json.Unmarshal(detailResponse, &clusterDetail); err != nil {
		Errorln(err.Error())
	} else {
		summary.Overview.Config.Name = clusterDetail.Name
		summary.Overview.Detail = clusterDetail
	}

	clusterNetwork := ClusterNetwork{}
	//fmt.Println(string(networkResponse))
	if err := json.Unmarshal(networkResponse, &clusterNetwork); err != nil {
		Errorln(err.Error())
	} else {
		if clusterNetwork.ClusterMgmtIpAddress.Fqdn != "" {
			summary.Overview.Config.MgmtIp.Addr = clusterNetwork.ClusterMgmtIpAddress.Fqdn
		} else {
			summary.Overview.Config.MgmtIp.Addr = clusterNetwork.ClusterMgmtIpAddress.Ip
		}
	}

	clusterHealth := ClusterHealth{}
	if err = json.Unmarshal(healthResponse, &clusterHealth); err != nil {
		Errorln(err.Error())
	} else {
		summary.Overview.Health = clusterHealth
	}

	clusterStats := ClusterStats{}
	if err = json.Unmarshal(statsResponse, &clusterStats); err != nil {
		Errorln(err.Error())
	} else {
		summary.Overview.Stats = clusterStats
	}

	clusterTime := ClusterTime{}
	if err = json.Unmarshal(timeResponse, &clusterTime); err != nil {
		Errorln(err.Error())
	} else {
		summary.Overview.Time = clusterTime
	}

	return &summary, nil
}

func getStorageCapacityInTB(capacity int) string {
	c := float64(capacity) / (1024 * 1024 * 1024 * 1024)
	return fmt.Sprintf("%.2fTB", c)
}

func getUptimeString(secs int64) string {
	days := secs / 86400
	secs = secs - (days * 86400)

	hours := (secs / 3600) % 24
	secs = secs - (hours * 3600)

	minutes := (secs / 60) % 60
	secs = secs - minutes*60

	seconds := secs % 60
	return fmt.Sprintf("%d Days,%d Hours,%d Minutes,%d Seconds", days, hours, minutes, seconds)
}
