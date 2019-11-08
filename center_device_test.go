package main

import (
	"testing"
	"time"
)

func TestFindBySerial(t *testing.T) {
	l := CenterDeviceList{
		&CenterDevice{1, CenterDeviceSerial{"a", "1"}, "", time.Now()},
		&CenterDevice{2, CenterDeviceSerial{"b", "2"}, "", time.Now()},
	}

	got := l.FindBySerial(&CenterDeviceSerial{"c", "2"})
	if got == nil {
		t.Error("FindBySerial(2) = nil, expected not nil")
	}
	if got.Num != 2 {
		t.Errorf("FindBySerial(2).Num = %d, expected %d", got.Num, 2)
	}

	got = l.FindBySerial(&CenterDeviceSerial{"c", "3"})
	if got != nil {
		t.Error("FindBySerial(3) <> nil, expected nil")
	}
}
