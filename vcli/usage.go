package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func printUsage() {
	prog := filepath.Base(os.Args[0])
	fmt.Println("Usage: \t", prog, "-h <ESXi or vCenter host> -u <Username> -p <Password>")
	os.Exit(1)
}
