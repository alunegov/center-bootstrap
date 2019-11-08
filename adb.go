package main

import (
	"errors"
	"os/exec"
	"path"
)

// Adb is a primitive adb wrapper.
type Adb struct {
	cmdName string
}

// NewAdb creates new Adb.
func NewAdb(adbPath string) *Adb {
	return &Adb{
		cmdName: path.Join(adbPath, "adb"),
	}
}

// CmdName return full name for adb executable.
func (it *Adb) CmdName() string {
	return it.cmdName
}

// RunCmd runs adb command, waits for it completition and returns combined stdout and stderr output. Adds "-d - directs
// command to the only connected USB device" to run command only on connected device, not on emulators and etc.
func (it *Adb) RunCmd(arg ...string) ([]byte, error) {
	if len(arg) == 0 {
		return nil, errors.New("empty arg")
	}
	arg_ := []string{"-d"}
	arg_ = append(arg_, arg...)
	cmd := exec.Command(it.cmdName, arg_...)
	return cmd.CombinedOutput()
}
