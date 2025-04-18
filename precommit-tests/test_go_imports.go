package main

import (
    "os"
    "fmt"
    "strings"
)

// This file has unorganized imports to test go-imports hook
func main() {
    fmt.Println("Testing imports")
    os.Exit(0)
    strings.ToLower("test")
}