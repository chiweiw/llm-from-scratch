package main

import (
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
)

func main() {
	err := wails.Run(&options.App{
		Title:  "Deploy Client",
		Width:  960,
		Height: 640,
		Assets: assets,
	})
	if err != nil {
		println(err.Error())
	}
}
