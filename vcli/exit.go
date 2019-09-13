package main

import (
	"fmt"
	"os"
)

type ExitCommand struct{}

func (c *ExitCommand) Execute(v *Vcli, args ...string) error {
	fmt.Println("Good Bye!")
	os.Exit(0)
	return nil
}
