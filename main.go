package main

import (
	"fmt"
	"os"
)

func main() {
	flags := parseFlags()

	if err := validateFlags(flags); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	runCLI(flags)
}
