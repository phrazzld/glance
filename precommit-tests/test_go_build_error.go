package main

// This file has build errors to test go-build hook
func main() {
    undefinedFunction() // Call to undefined function
}