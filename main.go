package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"
)

func main() {
	switch len(os.Args) {
	case 1:
		readNotes()
	case 2:
		writeNote(os.Args[1])
	default:
		log.Fatal("Too many args")
	}
}

func readNotes() {
	fmt.Println("Grand terminal app...")
}

func writeNote(key string) {

	t := time.Now()

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("Error:", err)
	}

	fpath := path.Join(home, ".notes", key)

	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		os.MkdirAll(fpath, 0700)
	}

	fname := filepath.Join(fpath, fmt.Sprintf("%s.md", t.Format("20060102150405")))

	f, err := os.Create(fname)
	if err != nil {
		log.Fatalln("Error:", err)
	}

	defer f.Close()

	e := os.Getenv("EDITOR")

	if e == "" {
		e = "vim"
	}

	cmd := exec.Command(e, fname)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err = cmd.Run()

	if err != nil {
		log.Fatalln("Error:", err)
	}

	fmt.Printf("Created: %s\n", fname)
}
