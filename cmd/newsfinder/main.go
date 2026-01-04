package main

import (
	"NewsFinder/internal/app"
)

func main() {
	nf := app.InitApp()

	nf.StartApp()

	// TODO: graceful shutdown, modular restarts, data and config autoupdates
	select {}
}
