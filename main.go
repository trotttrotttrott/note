package main

import (
	"log"
	"os"
	"path"
)

var notesDir string

func main() {

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("Error:", err)
	}

	notesDir = path.Join(home, ".notes")

	switch len(os.Args) {
	case 1:
		read()
	case 2:
		write(os.Args[1])
	default:
		log.Fatal("Too many args")
	}
}
