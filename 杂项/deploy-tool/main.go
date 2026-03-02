package main

import (
	appcmd "deploy-tool/cmd/app"
	"embed"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if err := appcmd.Run(assets); err != nil {
		println("Error:", err.Error())
	}
}
