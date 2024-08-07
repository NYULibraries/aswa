package main

import "log"

func main() {
	err := DoCheck()
	if err != nil {
		log.Fatal("Error:", err)
	}
}
