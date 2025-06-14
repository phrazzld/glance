package main

import (
	"fmt"

	"golang.org/x/text/language"
)

func main() {
	// Simple usage of a clean golang.org/x/text package
	// This package version (v0.14.0) should have no known vulnerabilities
	tag := language.English
	fmt.Printf("Language tag: %s\n", tag)
}
