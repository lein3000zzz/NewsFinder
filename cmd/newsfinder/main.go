package main

import (
	"NewsFinder/internal/app"
)

func main() {
	nf := app.InitApp()

	nf.StartApp()

	select {}
}
