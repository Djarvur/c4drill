// Package main provides the CLI entry point for c4drill.
package main

import (
	"os"
)

func main() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
	// Exit 0 is implicit
}
