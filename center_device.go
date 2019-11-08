package main

import (
	"time"
)

// CenterDeviceSerial is a serial number of Center device.
type CenterDeviceSerial struct {
	Key   string
	Value string
}

// CenterDevice is a Center device.
type CenterDevice struct {
	Num       int
	Serial    CenterDeviceSerial
	AndroidId string
	AddTime   time.Time
}

// CenterDeviceList is a list of Center devices.
type CenterDeviceList []*CenterDevice

// FindBySerial finds Center device by its serial number. Returns nil if no device found.
func (it CenterDeviceList) FindBySerial(serial *CenterDeviceSerial) *CenterDevice {
	for _, centerDevice := range it {
		if centerDevice.Serial.Value == serial.Value {
			return centerDevice
		}
	}
	return nil
}
