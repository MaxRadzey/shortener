package main

import "github.com/MaxRadzey/shortener/internal/app"

func main() {
	if err := app.Run(); err != nil {
		panic(err)
	}
}
