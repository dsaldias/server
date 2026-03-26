package main

import (
	"os"
	"strings"
)

func getModuleName() string {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		panic("no go.mod found")
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}

	panic("module not found")
}
