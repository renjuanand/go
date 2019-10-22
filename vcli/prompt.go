package main

import (
	prompt "github.com/c-bata/go-prompt"
)

func showPrompt() {
	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix("==> "),
		prompt.OptionTitle("vcli"),
		// prompt.OptionPrefixTextColor(prompt.Turquoise),
		// prompt.OptionPrefixTextColor(prompt.Fuchsia),
		prompt.OptionPrefixTextColor(prompt.Yellow),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionInputTextColor(prompt.Green),
		// prompt.OptionHistory([]string{"about", "version", "vm list"}),
		prompt.OptionSuggestionBGColor(prompt.DarkGray))
	p.Run()
}
