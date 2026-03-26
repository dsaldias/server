package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("expected command: init")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		Init()
	default:
		fmt.Println("unknown command")
	}
}
