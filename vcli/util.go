package main

import (
	"github.com/fatih/color"
)

var (
	Error   = color.New(color.FgRed).PrintfFunc()
	Warn    = color.New(color.FgYellow).PrintFunc()
	Info    = color.New(color.FgBlue).PrintFunc()
	Title   = color.New(color.Bold, color.FgCyan).SprintFunc()
	Notice  = color.New(color.Bold, color.FgGreen).PrintlnFunc()
	Message = color.New(color.Bold, color.FgHiWhite).PrintlnFunc()
	Success = color.New(color.FgGreen).PrintlnFunc()
)
