package main

import (
	"log"
	"os"
)

func main() {
	switch len(os.Args) {
	case 1:
		read()
	case 2:
		write(os.Args[1])
	default:
		log.Fatal("Too many args")
	}
}
