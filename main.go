package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"time"
)

type CenterDeviceSerial struct {
	Key   string
	Value string
}

type CenterDevice struct {
	Num       int
	Serial    CenterDeviceSerial
	AndroidId string
	AddTime   time.Time
}

type CenterDeviceList []*CenterDevice

const AdbFilePath = `c:/dev/android-sdk/platform-tools/adb`
const CenterDevicesDb = `./CenterDevices.json`
const CenterVersion = "1.5"
const LogsPath = `./_log`
const ApkOutputPath = `./app/build/outputs/apk`
const DeviceApkPath = `/sdcard/Download/`

func main() {
	fmt.Printf("Loading %s...\n", CenterDevicesDb)
	centerDevices, err := centerDevice_FindAll()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Getting device propertys...")
	cmd := exec.Command(AdbFilePath, "shell", "getprop")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s\n", output)
		fmt.Println(err)
		return
	}

	fmt.Println("Detecting device serial...")
	serials := make([]*CenterDeviceSerial, 0)
	re := regexp.MustCompile(`\[(.+)\]: \[(.+)\]`)
	reader := bytes.NewReader(output)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.Contains(text, "serial") {
			a := re.FindStringSubmatch(text)
			if a == nil {
				continue
			}

			serials = append(serials, &CenterDeviceSerial{a[1], a[2]})

			fmt.Printf("%d. %s: %s\n", len(serials), a[1], a[2])
		}
	}

	fmt.Print("Select suitable serial: ")
	var selectedSerialNum int
	n, err := fmt.Scanf("%d\n", &selectedSerialNum)
	if err != nil {
		fmt.Println(err)
		return
	}
	if (n != 1) || (selectedSerialNum == 0) || (selectedSerialNum > len(serials)) {
		fmt.Println("Nothing/wrong selected")
		return
	}

	fmt.Print("ANDROID_ID: ")
	androidId := ""
	_, _ = fmt.Scanf("%s\n", &androidId)

	selectedSerial := serials[selectedSerialNum-1]
	centerDevice := centerDevice_FindBySerial(centerDevices, selectedSerial)
	if centerDevice != nil {
		fmt.Printf("Rebuild Center #%d? ", centerDevice.Num)
		var choice string
		if _, err := fmt.Scanf("%s\n", &choice); (err != nil) || (choice != "y") {
			fmt.Println("Nothing/wrong selected")
			return
		}

		if androidId != "" {
			centerDevice.AndroidId = androidId
		}
	} else {
		nextNum := 0
		if len(centerDevices) == 0 {
			nextNum = 344
		} else {
			nextNum = centerDevices[len(centerDevices)-1].Num + 1
		}

		fmt.Printf("New Center #%d? ", nextNum)
		var newNextNum int
		if n, err := fmt.Scanf("%d\n", &newNextNum); (err == nil) && (n == 1) {
			nextNum = newNextNum
		}

		centerDevice = &CenterDevice{nextNum, *selectedSerial, androidId, time.Now()}

		centerDevices = append(centerDevices, centerDevice)
	}

	fmt.Printf("Saving %s...\n", CenterDevicesDb)
	if err := centerDevice_Store(centerDevices); err != nil {
		fmt.Println(err)
		return
	}

	logFilePath := path.Join(path.Join(LogsPath, fmt.Sprintf(`C%d.txt`, centerDevice.Num)))
	if _, err := os.Stat(logFilePath); err != nil {
		fmt.Printf("Saving device propertys to %s...\n", logFilePath)
		err = os.MkdirAll(LogsPath, 0)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = ioutil.WriteFile(logFilePath, output, 0)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	fmt.Println("Building apk...")
	if centerDevice.AndroidId == "" {
		cmd = exec.Command("gradlew",
			"-PserialKey="+centerDevice.Serial.Key,
			"-PserialValue="+centerDevice.Serial.Value,
			"assembleRelease")
	} else {
		cmd = exec.Command("gradlew",
			"-PserialKey=ANDROID_ID",
			"-PserialValue="+centerDevice.AndroidId,
			"assembleRelease")
	}
	output, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s\n", output)
		fmt.Println(err)
		return
	}

	apkName := fmt.Sprintf(`Center-%d_%s.apk`, centerDevice.Num, CenterVersion)

	fmt.Printf("Renaming app-release.apk to %s...\n", apkName)
	err = os.Rename(path.Join(ApkOutputPath, `app-release.apk`), apkName)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Installing apk to device...")
	cmd = exec.Command(AdbFilePath, "install", "-rg", apkName)
	output, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s\n", output)
		fmt.Println(err)
		return
	}

	fmt.Printf("Copying apk to device %s...\n", DeviceApkPath)
	cmd = exec.Command(AdbFilePath, "push", apkName, DeviceApkPath)
	output, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s\n", output)
		fmt.Println(err)
		return
	}
}

func centerDevice_FindAll() (CenterDeviceList, error) {
	if _, err := os.Stat(CenterDevicesDb); err != nil {
		return make(CenterDeviceList, 0), nil
	}

	data, err := ioutil.ReadFile(CenterDevicesDb)
	if err != nil {
		return nil, err
	}

	var res CenterDeviceList

	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}

	return res, nil
}

func centerDevice_Store(list CenterDeviceList) error {
	data, err := json.MarshalIndent(list, "", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(CenterDevicesDb, data, 0)
}

func centerDevice_FindBySerial(list CenterDeviceList, serial *CenterDeviceSerial) *CenterDevice {
	for _, centerDevice := range list {
		if centerDevice.Serial.Value == serial.Value {
			return centerDevice
		}
	}
	return nil
}
