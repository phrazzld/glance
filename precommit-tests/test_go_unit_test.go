package main

import "testing"

// This file has a failing test to trigger go-unit-tests hook
func TestThatFails(t *testing.T) {
    t.Error("This test intentionally fails")
}