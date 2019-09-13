package main

type Command interface {
	Execute(v *Vcli, args ...string) error
}

// commands available for vCLI prompt
var Commands = map[string]Command{
	"about":   &AboutCommand{},
	"version": &VersionCommand{},
	"vm":      &VmCommand{},
	"exit":    &ExitCommand{},
	"quit":    &ExitCommand{},
}
