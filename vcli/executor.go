package main

import (
	"fmt"
	"strings"
)

func executor(command string) {
	vcli := GetVcli()
	cmds := strings.Split(strings.Trim(command, " "), " ")

	if len(command) > 0 && len(cmds) > 0 {
		pCmd := cmds[0]
		options := cmds[1:]
		if fn, ok := Commands[pCmd]; ok {
			err := fn.Execute(vcli, options...)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Printf("Unknown command: '%s'\n", pCmd)
		}
	}
}
