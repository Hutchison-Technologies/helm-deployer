package main

import "testing"

func TestMain(t *testing.T) {
	want := "Main, world."
	if got := Hello(); got != want {
		t.Errorf("Main() = %q, want %q", got, want)
	}
}
