package main

// This file intentionally has bad formatting to test go-fmt hook
func main() {
fmt.Println("Hello World")
    if true {
        fmt.Println("Indented incorrectly")
}
}