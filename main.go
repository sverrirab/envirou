package main

import (
	_ "embed"
	"github.com/sverrirab/envirou/cmd"
)

// These variables contain embedded scripts
//
//go:embed bash/ev.sh
var embeddedBootstrapBash string

//go:embed bash/prompt.bash
var embeddedPromptBash string

//go:embed bash/prompt.zsh
var embeddedPromptZsh string

//go:embed powershell/ev.ps1
var embeddedBootstrapPowerShell string

//go:embed powershell/prompt.ps1
var embeddedPromptPowerShell string

//go:embed ev.cmd
var embeddedBootstrapBat string

func main() {
	cmd.Execute(embeddedBootstrapBash, embeddedPromptBash, embeddedPromptZsh, embeddedBootstrapPowerShell, embeddedPromptPowerShell, embeddedBootstrapBat)
}
