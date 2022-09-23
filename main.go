// Package main provides ...
package main

import (
	"github.com/atotto/clipboard"
)

func main() {
	// fmt.Println("Hello World")
	content, err := clipboard.ReadAll()
	if err != nil {
		panic(err)
	}
	println(content)
}

