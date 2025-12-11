package main

import (
	"log"

	"github.com/NYULibraries/aswa/cmd"
)

func main() {
	err := cmd.DoCheck()
	if err != nil {
		log.Fatal("Error:", err)
	}
}
