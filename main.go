package main

import (
	_ "embed"
	"github.com/sverrirab/envirou/cmd"
)

// These variables contain embedded scripts
//
//go:embed bash/ev.sh
var embeddedBootstrapBash string

//go:embed powershell/ev.ps1
var embeddedBootstrapPowerShell string

//go:embed ev.cmd
var embeddedBootstrapBat string

func main() {
	cmd.Execute(embeddedBootstrapBash, embeddedBootstrapPowerShell, embeddedBootstrapBat)
}
