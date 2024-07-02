package main

import "github.com/leodayo/url-shortener/internal/app"

func main() {
	err := app.Run()

	if err != nil {
		panic(err)
	}
}
