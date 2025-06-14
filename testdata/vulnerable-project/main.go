package main

import (
	"fmt"

	_ "golang.org/x/crypto/ssh" // Import for vulnerability testing
	"golang.org/x/text/language"
)

func main() {
	// Simple usage of vulnerable packages for testing vulnerability detection
	// golang.org/x/text v0.3.0 - may contain CVE-2020-14040
	// golang.org/x/crypto from 2019 - likely contains various CVEs

	tag := language.English
	fmt.Printf("Language tag: %s\n", tag)

	// Reference crypto package (without actually using vulnerable functions)
	// Just import it to trigger vulnerability scanning
	fmt.Println("Referenced crypto package for vulnerability testing")
}
