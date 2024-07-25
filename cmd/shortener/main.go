package main

import (
	"fmt"

	"github.com/leodayo/url-shortener/internal/app"
)

func main() {
	err := app.Run()

	if err != nil {
		fmt.Println(err)
	}
}
