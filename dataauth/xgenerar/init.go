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

	lines := strings.SplitSeq(string(data), "\n")
	for line := range lines {
		if after, ok := strings.CutPrefix(line, "module "); ok {
			return strings.TrimSpace(after)
		}
	}

	panic("module not found")
}
