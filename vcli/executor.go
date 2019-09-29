package main

import (
	"strings"
)

func executor(command string) {
	vcli := GetVcli()
	cmds := strings.Split(strings.Trim(command, " "), " ")

	if len(command) > 0 && len(cmds) > 0 {
		pCmd := cmds[0]
		options := cmds[1:]

		if fn, ok := Commands[pCmd]; ok {
			// Start spinner before executing the command
			Spinner.Start()
			t, err := fn.Execute(vcli, options...)
			// Stop spinner once command execution is finished
			Spinner.Stop()

			if err != nil {
				Error(err.Error())
			}
			// Print command response
			if t != nil {
				t.Print()
			}
		} else {
			Error("Unknown command: '%s'\n", pCmd)
		}
	}
}
