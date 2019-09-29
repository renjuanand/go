package main

import (
	"github.com/tatsushid/go-prettytable"
	"os"
)

type ExitCommand struct{}

func (c *ExitCommand) Execute(v *Vcli, args ...string) (*prettytable.Table, error) {
	// Exiting the application, stop spinner
	if Spinner.Active() {
		Spinner.Stop()
	}
	Message("Good Bye!")
	os.Exit(0)
	return nil, nil
}
