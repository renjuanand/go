package main

import (
	"context"
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"

	"github.com/vmware/govmomi"
	_ "github.com/vmware/govmomi/vim25"
	_ "github.com/vmware/govmomi/vim25/mo"
)

type Vcli struct {
	ctx     context.Context
	client  *govmomi.Client
	channel chan string
}

var (
	once     sync.Once
	instance *Vcli
)

var protocolMatch = regexp.MustCompile(`^\w+://`)

func New(url string, username string, passwd string, insecure bool) (*Vcli, error) {
	if instance == nil {
		u, err := getURL(url, username, passwd)
		if err != nil {
			return nil, err
		}
		ctx := context.Background()
		c, err := govmomi.NewClient(ctx, u, insecure)
		if err != nil {
			return nil, err
		}

		messages := make(chan string)
		// fmt.Printf("Client: %v\n", c)
		once.Do(func() {
			instance = &Vcli{
				ctx:     ctx,
				client:  c,
				channel: messages,
			}
		})
	}
	return instance, nil
}

func GetVcli() *Vcli {
	return instance
}

func getURL(host string, user string, password string) (*url.URL, error) {
	var err error
	var u *url.URL

	if host != "" {
		// Default protocol prefix to https
		if !protocolMatch.MatchString(host) {
			host = "https://" + host
		}

		u, err = url.Parse(host)
		if err != nil {
			return nil, err
		}

		// Default the path to /sdk
		if u.Path == "" {
			u.Path = "/sdk"
		}

		if u.User == nil {
			u.User = url.UserPassword(user, password)
		}
	}

	return u, nil
}

func getArgs() (string, string, string, bool) {
	vcliArgs := flag.NewFlagSet(filepath.Base(os.Args[0]), flag.ExitOnError)
	url := vcliArgs.String("h", "", "ESXi or vCenter host")
	username := vcliArgs.String("u", "", "Username")
	password := vcliArgs.String("p", "", "Password")
	insecure := vcliArgs.Bool("k", true, "Insecure")

	if len(os.Args) <= 1 {
		printUsage()
	}

	if strings.Trim(*password, " ") == "" {
		fmt.Print("Enter password: ")
		passwd, err := terminal.ReadPassword(int(syscall.Stdin))
		if err == nil {
			*password = string(passwd)
		}
		fmt.Println()
	}

	vcliArgs.Parse(os.Args[1:])
	return *url, *username, *password, *insecure
}

func main() {
	h, u, p, s := getArgs()

	if h == "" || u == "" || p == "" {
		printUsage()
	}

	cli, err := New(h, u, p, s)

	if err != nil {
		log.Fatal(err)
	}

	defer cli.client.Logout(cli.ctx)

	fmt.Println("vCLI - An interactive vSphere CLI client")

	// Show messages we recieve from other go routines
	go func() {
		for m := range cli.channel {
			fmt.Println(m)
		}
	}()

	showPrompt()
}
