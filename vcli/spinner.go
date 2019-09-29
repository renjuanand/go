package main

import (
	"github.com/briandowns/spinner"
	"time"
)

var Spinner = buildSpinner()

func buildSpinner() *spinner.Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Prefix = "Fetching data... "
	s.Color("white", "bold")
	s.Stop()
	return s
}
