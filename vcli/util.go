package main

import (
	"github.com/fatih/color"
	_ "github.com/mattn/go-colorable"
)

type KeyValue struct {
	key   string
	value string
}

var (
	Error   = color.New(color.FgRed).PrintfFunc()
	ErrorSp = color.New(color.FgRed).SprintFunc()
	Errorln = func(args ...interface{}) {
		if Spinner.Active() {
			Spinner.Stop()
			color.New(color.FgRed).Println(args...)
			Spinner.Start()
		} else {
			color.New(color.FgRed).Println(args...)
		}
	}

	Usage = func(args ...interface{}) {
		if Spinner.Active() {
			Spinner.Stop()
			color.New(color.FgCyan).Println(args...)
			Spinner.Start()
		} else {
			color.New(color.FgCyan).Println(args...)
		}
	}

	Warn    = color.New(color.FgYellow).PrintFunc()
	Warnln  = color.New(color.FgYellow).PrintlnFunc()
	Info    = color.New(color.FgHiWhite).PrintFunc()
	Infoln  = color.New(color.FgHiWhite).PrintlnFunc()
	InfoSp  = color.New(color.FgHiWhite).SprintFunc()
	Title   = color.New(color.Bold, color.FgCyan).SprintFunc()
	Notice  = color.New(color.Bold, color.FgGreen).PrintlnFunc()
	Message = color.New(color.Bold, color.FgHiWhite).PrintlnFunc()
	Key     = color.New(color.FgHiWhite, color.Bold).SprintFunc()
	Value   = color.New(color.FgWhite).SprintFunc()
	Success = color.New(color.FgGreen).PrintlnFunc()
	Debug   = color.New(color.FgWhite).PrintfFunc()
	Debugln = color.New(color.FgWhite).PrintlnFunc()
	HC      = color.New(color.Bold, color.FgHiCyan).PrintFunc()
	HD      = color.New(color.FgWhite).PrintlnFunc()
	HE      = color.New(color.FgHiWhite).SprintFunc()
)
