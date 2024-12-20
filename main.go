package main

import (
	"github.com/NYULibraries/aswa/cmd"
	"log"
)

func main() {
	err := cmd.DoCheck()
	if err != nil {
		log.Fatal("Error:", err)
	}
}
