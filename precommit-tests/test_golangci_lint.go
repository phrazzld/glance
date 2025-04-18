package main

import "fmt"

// This file has linting issues to test golangci-lint hook
func main() {
    var unused string // Unused variable
    if true { fmt.Println("Unreachable code") }
}