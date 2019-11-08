package main

import (
	"path"
	"testing"
)

func TestCmdName(t *testing.T) {
	adb := NewAdb("test")

	exp := path.Join("test", "adb")
	got := adb.CmdName()
	if got != exp {
		t.Errorf("CmdName() = %s, expected %s", got, exp)
	}
}
