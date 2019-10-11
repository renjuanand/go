package main

import (
	"github.com/fatih/color"
)

type KeyValue struct {
	key   string
	value string
}

var (
	Error    = color.New(color.FgRed).PrintfFunc()
	ErrorStr = color.New(color.FgRed).SprintFunc()
	Errorln  = color.New(color.FgRed).PrintlnFunc()
	Warn     = color.New(color.FgYellow).PrintFunc()
	Warnln   = color.New(color.FgYellow).PrintlnFunc()
	Info     = color.New(color.FgBlue).PrintFunc()
	Infoln   = color.New(color.FgBlue).PrintlnFunc()
	InfoStr  = color.New(color.FgBlue).SprintFunc()
	Title    = color.New(color.Bold, color.FgCyan).SprintFunc()
	Notice   = color.New(color.Bold, color.FgGreen).PrintlnFunc()
	Message  = color.New(color.Bold, color.FgHiWhite).PrintlnFunc()
	Key      = color.New(color.FgHiWhite, color.Bold).SprintFunc()
	Value    = color.New(color.FgWhite).SprintFunc()
	Success  = color.New(color.FgGreen).PrintlnFunc()
	Debug    = color.New(color.FgWhite).PrintfFunc()
	Debugln  = color.New(color.FgWhite).PrintlnFunc()
	HC       = color.New(color.FgHiGreen).SprintFunc()
	HD       = color.New(color.FgWhite).SprintFunc()
	HE       = color.New(color.FgHiWhite).SprintFunc()
)
