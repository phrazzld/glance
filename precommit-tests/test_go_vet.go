package main

import "fmt"

// This file has suspicious code patterns to test go-vet hook
func main() {
    var x int
    fmt.Printf("%d %d", x) // Wrong number of arguments
}