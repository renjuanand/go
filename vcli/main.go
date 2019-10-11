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
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/vmware/govmomi"
)

type Credentials struct {
	username string
	password string
}

type Vcli struct {
	ctx    context.Context
	client *govmomi.Client
	auth   *Credentials
}

type Exit int

const (
	VCLI_VERSION        = "1.0.0"
	IDLE_ACTION_TIMEOUT = 5 * time.Second
	SessionCookieName   = "vmware_soap_session"
)

var (
	once     sync.Once
	instance *Vcli
)

var IdleActionTimer *time.Timer
var protocolMatch = regexp.MustCompile(`^\w+://`)

// show vCLI usage
func printUsage() {
	prog := filepath.Base(os.Args[0])
	fmt.Println("Usage: \t", prog, "-h <ESXi or vCenter host> -u <Username> -p <Password>")
	os.Exit(1)
}

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

		/*
			for _, cookie := range c.Client.Client.Client.Jar.Cookies(u) {
				if cookie.Name == SessionCookieName {
					fmt.Println("soap session cookie: ", cookie.Value)
				}
			}
		*/
		once.Do(func() {
			instance = &Vcli{
				ctx:    ctx,
				client: c,
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
	// insecure := vcliArgs.Bool("k", true, "Insecure")
	// Don't verify the server's certificate chain (default)
	insecure := true

	if len(os.Args) <= 1 {
		printUsage()
	}

	vcliArgs.Parse(os.Args[1:])

	if strings.Trim(*password, " ") == "" {
		fmt.Print("Enter password: ")
		passwd, err := terminal.ReadPassword(int(syscall.Stdin))
		if err == nil {
			*password = string(passwd)
		}
		fmt.Println()
	}

	return *url, *username, *password, insecure
}

func handleExit(cli *Vcli) {
	switch v := recover().(type) {
	case nil:
		cli.client.Logout(cli.ctx)
		Message("Good Bye!")
		return
	case Exit:
		cli.client.Logout(cli.ctx)
		os.Exit(int(v))
	default:
		fmt.Println(string(debug.Stack()))
	}
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
	auth := &Credentials{username: u, password: p}
	cli.auth = auth

	/*
		IdleActionTimer = time.NewTimer(IDLE_ACTION_TIMEOUT)

		go func() {
			<-IdleActionTimer.C
			cli.client.Logout(cli.ctx)
			Message("Session disconnected!")
			panic(Exit(0))
		}()
	*/
	defer handleExit(cli)
	defer cli.client.Logout(cli.ctx)

	a := cli.client.Client.ServiceContent.About
	Success("Connected to", a.Name, a.Version)
	showPrompt()
}
