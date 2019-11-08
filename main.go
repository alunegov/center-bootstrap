package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"time"
)

const (
	adbFilePath         = `C:/dev/android-sdk/platform-tools`
	centerDevicesDb     = `CenterDevices.json`
	logsPath            = `_log`
	gradleBuildFileName = `app/build.gradle`
	apkOutputPath       = `app/build/outputs/apk`
	deviceApkPath       = `/sdcard/Download/`
	centerIDApk         = "_id/center-id.apk"
	centerIDPkg         = "ru.ros_diagnostics.centerid"
	centerIDTag         = "CenterId"
)

func main() {
	adb := NewAdb(adbFilePath)

	fmt.Println("Getting device propertys...")
	deviceProps, err := adb.RunCmd("shell", "getprop")
	if err != nil {
		fmt.Printf("%s\n", deviceProps)
		fmt.Println(err)
		return
	}

	fmt.Println("Detecting device serial...")
	serials := extractSerials(deviceProps)
	for i, serial := range serials {
		fmt.Printf("%d. %s: %s\n", i+1, serial.Key, serial.Value)
	}

	fmt.Print("Select suitable serial: ")
	selectedSerialNum, err := readInt()
	if err != nil {
		fmt.Println(err)
		return
	}
	if (selectedSerialNum < 1) || (len(serials) < selectedSerialNum) {
		fmt.Println("Empty/wrong input")
		return
	}
	selectedSerial := serials[selectedSerialNum-1]

	fmt.Print("Use ANDROID_ID (only for Android 8 and higher)? (a/m/n) ")
	useAndroidID, err := readString()
	if err != nil {
		fmt.Println(err)
		return
	}

	var androidID string
	if useAndroidID == "a" {
		fmt.Println("Detecting device ANDROID_ID...")
		var output []byte
		androidID, output, err = detectAndroidID(adb)
		if err != nil {
			fmt.Printf("%s\n", output)
			fmt.Println(err)
			return
		}
		fmt.Printf("ANDROID_ID = %s\n", androidID)
	} else if useAndroidID == "m" {
		fmt.Print("Enter ANDROID_ID: ")
		androidID, err = readString()
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		androidID = ""
	}

	fmt.Printf("Loading %s...\n", centerDevicesDb)
	centerDevices, err := centerDeviceFindAll()
	if err != nil {
		fmt.Println(err)
		return
	}

	centerDevice := centerDevices.FindBySerial(selectedSerial)
	if centerDevice != nil {
		fmt.Printf("Rebuild Center #%d? (y/n) ", centerDevice.Num)
		choice, err := readString()
		if (err != nil) || (choice != "y") {
			fmt.Println("Empty/wrong input or exit")
			return
		}

		if androidID != "" {
			centerDevice.AndroidId = androidID
		}
	} else {
		nextNum := 0
		if len(centerDevices) == 0 {
			nextNum = 344
		} else {
			nextNum = centerDevices[len(centerDevices)-1].Num + 1
		}

		fmt.Printf("New Center #%d? (you can enter different #) ", nextNum)
		newNextNum, err := readInt()
		if err == nil {
			nextNum = newNextNum
		}

		centerDevice = &CenterDevice{nextNum, *selectedSerial, androidID, time.Now()}

		centerDevices = append(centerDevices, centerDevice)
	}

	fmt.Printf("Saving %s...\n", centerDevicesDb)
	if err := centerDeviceStore(centerDevices); err != nil {
		fmt.Println(err)
		return
	}
	logFilePath := path.Join(logsPath, fmt.Sprintf(`C%d.txt`, centerDevice.Num))
	if _, err := os.Stat(logFilePath); err != nil {
		fmt.Printf("Saving device propertys to %s...\n", logFilePath)
		err = os.MkdirAll(logsPath, 0)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = ioutil.WriteFile(logFilePath, deviceProps, 0)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	fmt.Println("Building apk...")
	var serialKeyParam, serialValueParam string
	if centerDevice.AndroidId == "" {
		serialKeyParam = "-PserialKey=" + centerDevice.Serial.Key
		serialValueParam = "-PserialValue=" + centerDevice.Serial.Value
	} else {
		serialKeyParam = "-PserialKey=ANDROID_ID"
		serialValueParam = "-PserialValue=" + centerDevice.AndroidId
	}
	cmd := exec.Command("gradlew", serialKeyParam, serialValueParam, "assembleRelease")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s\n", output)
		fmt.Println(err)
		return
	}

	fmt.Println("Detecting apk version...")
	version, err := detectVersion(gradleBuildFileName)
	if err != nil {
		fmt.Println(err)
		return
	}

	apkName := fmt.Sprintf(`Center-%d_%s.apk`, centerDevice.Num, version)

	fmt.Printf("Renaming app-release.apk to %s...\n", apkName)
	err = os.Rename(path.Join(apkOutputPath, `app-release.apk`), apkName)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Installing apk to device...")
	output, err = adb.RunCmd("install", "-rg", apkName)
	if err != nil {
		fmt.Printf("%s\n", output)
		fmt.Println(err)
		return
	}

	fmt.Printf("Copying apk to device %s...\n", deviceApkPath)
	output, err = adb.RunCmd("push", apkName, deviceApkPath)
	if err != nil {
		fmt.Printf("%s\n", output)
		fmt.Println(err)
		return
	}
}

// readString reads string from stdin.
func readString() (string, error) {
	var res string
	n, err := fmt.Scanf("%s\n", &res)
	if err != nil {
		return "", err
	}
	if n != 1 {
		return "", fmt.Errorf("Scanf expected 1 item but got %d items", n)
	}
	return res, nil
}

// readString reads integer from stdin.
func readInt() (int, error) {
	var res int
	n, err := fmt.Scanf("%d\n", &res)
	if err != nil {
		return 0, err
	}
	if n != 1 {
		return 0, fmt.Errorf("Scanf expected 1 item but got %d items", n)
	}
	return res, nil
}

// extractSerials extracts serials from captured output of "adb shell getprop" by searching strings with text "serial".
func extractSerials(output []byte) []*CenterDeviceSerial {
	serials := make([]*CenterDeviceSerial, 0)
	// из строки [gsm.serial]: [VPR081831768                                                10P] выдираем пару
	// gsm.serial - VPR081831768. Часть после пробела в серийнике отбрасываем - так делали раньше вручную, и проверка
	// в Центровке ориентируется на для длину reference-строки.
	extract(bytes.NewReader(output), `\[(.*serial.*)\]: \[(\S+).*\]`, func(s []string) {
		serials = append(serials, &CenterDeviceSerial{s[1], s[2]})
	})
	return serials
}

// centerDeviceFindAll loads all Center devices from centerDevicesDb
func centerDeviceFindAll() (CenterDeviceList, error) {
	if _, err := os.Stat(centerDevicesDb); err != nil {
		return make(CenterDeviceList, 0), nil
	}

	data, err := ioutil.ReadFile(centerDevicesDb)
	if err != nil {
		return nil, err
	}

	var res CenterDeviceList

	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}

	return res, nil
}

// centerDeviceStore saves Center devices to centerDevicesDb
func centerDeviceStore(list CenterDeviceList) error {
	data, err := json.MarshalIndent(list, "", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(centerDevicesDb, data, 0)
}

// detectAndroidID detects ANDROID_ID by installing and running centerIdApk on device. centerIdApk output captured by
// running "adb logcat" in parallel.
func detectAndroidID(adb *Adb) (id string, output []byte, err error) {
	output, err = adb.RunCmd("install", "-rg", centerIDApk)
	if err != nil {
		return
	}
	defer func() {
		output, err = adb.RunCmd("uninstall", centerIDPkg)
	}()

	logcatCmd := exec.Command(adb.CmdName(), "-d", "logcat", "-s", fmt.Sprintf("%s:D", centerIDTag))
	var logcatOutput bytes.Buffer
	logcatCmd.Stdout = &logcatOutput
	err = logcatCmd.Start()
	if err != nil {
		return "", nil, err
	}

	// will wait for activity to start because of "-W"
	output, err = adb.RunCmd("shell", "am", "start", "-W", "-a android.intent.action.MAIN", fmt.Sprintf("-n %s/.MainActivity", centerIDPkg))
	if err != nil {
		return
	}
	// TODO: sleep?
	output, err = adb.RunCmd("shell", "am", "force-stop", centerIDPkg)
	if err != nil {
		return
	}

	// Process.Signal(os.Interrupt) don't work on Windows
	err = logcatCmd.Process.Kill()
	if err != nil {
		return "", nil, err
	}
	err = logcatCmd.Wait()
	if err != nil {
		// Wait can return error in case of process kill
		if _, ok := err.(*exec.ExitError); !ok {
			return "", nil, err
		}
	}

	id = extractAndroidID(logcatOutput.Bytes())
	if id == "" {
		return "", nil, errors.New("ANDROID_ID not detected")
	}

	return id, nil, nil
}

// extractAndroidID extracts ANDROID_ID from captured output of centerIdApk by searching string with text centerIDTag.
// Message with "id1" is a system prop "ro.serialno", message with "id2" - ANDROID_ID.
func extractAndroidID(output []byte) string {
	res := ""
	extract(bytes.NewReader(output), fmt.Sprintf(`.*%s.*: id2 = (.+)`, centerIDTag), func(s []string) {
		res = s[1]
	})
	return res
}

// detectVersion extracts version from build.gradle simply by parsing it for versionName field.
func detectVersion(gradleBuildFileName string) (string, error) {
	f, err := os.Open(gradleBuildFileName)
	if err != nil {
		return "", err
	}
	defer func(f_ *os.File) {
		if err2 := f_.Close(); err2 != nil {
			log.Printf("detectVersion got error while closing %s: %v\n", gradleBuildFileName, err2)
		}
	}(f)

	version := extractVersion(f)

	return version, nil
}

// extractVersion extracts apk version from content of app/build.gradle.
func extractVersion(r io.Reader) string {
	res := ""
	extract(r, `\s*versionName\s*['"](.+)['"]`, func(s []string) {
		res = s[1]
	})
	return res
}

// extract scans r line-by-line and try to find regexp constructed from str. For every str found calls cb.
func extract(r io.Reader, str string, cb func(s []string)) {
	re := regexp.MustCompile(str)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		text := scanner.Text()

		// somehow where is double CR in adb output
		text = dropCR(text)

		a := re.FindStringSubmatch(text)
		if a == nil {
			continue
		}

		cb(a)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("extract got error while scanning: %v\n", err)
	}
}

func dropCR(data string) string {
	return strings.TrimSuffix(data, "\r")
}
