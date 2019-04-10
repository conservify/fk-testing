package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type WpaSupplicantRunner struct {
	*BackgroundProcess
	ActiveSsid string
	Networks   []*WifiNetwork
	Device     string
}

func (wsr *WpaSupplicantRunner) Line(text string) {
	ssidRe := regexp.MustCompile("SSID='(\\S+)'")
	ssidM := ssidRe.FindStringSubmatch(text)
	if ssidM != nil {
		wsr.ActiveSsid = ssidM[0]
	}
	if strings.Contains(text, "CTRL-EVENT-CONNECTED") {
	}
	if strings.Contains(text, "CTRL-EVENT-DISCONNECTED") {
	}
}

func NewWpaSupplicantRunner(device string, networks []*WifiNetwork) (wsr *WpaSupplicantRunner, err error) {
	wsr = &WpaSupplicantRunner{
		Device:   device,
		Networks: networks,
	}

	argv := []string{"wpa_supplicant", "-D", "wext", "-i", device, "-c", "./wpa_supplicant.conf"}

	log.Printf("Executing %s", strings.Join(argv, " "))

	bp, err := NewBackgroundProcess("WPA  | ", argv, wsr)
	if err != nil {
		return
	}

	wsr.BackgroundProcess = bp

	wsr.writeWpaSupplicantConfig()

	return
}

func (wsr *WpaSupplicantRunner) writeWpaSupplicantConfig() error {
	f, err := os.OpenFile("wpa_supplicant.conf", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	w := bufio.NewWriter(f)

	defer w.Flush()

	w.WriteString("ctrl_interface=DIR=/run/wpa_supplicant_fk GROUP=adm\n")
	w.WriteString("update_config=1\n")

	for _, network := range wsr.Networks {
		w.WriteString("network={\n")
		w.WriteString(fmt.Sprintf("  ssid=\"%s\"\n", network.Name))
		w.WriteString(fmt.Sprintf("  priority=%d\n", network.Priority))
		if network.Password != "" {
			w.WriteString(fmt.Sprintf("  psk=\"%s\"\n", network.Password))
		} else {
			w.WriteString(fmt.Sprintf("  key_mgmt=NONE\n"))
		}
		w.WriteString("}\n")
	}

	return nil
}

type WpaCliRunner struct {
	Device    string
	WpaSocket string
}

type CurrentWifiState struct {
	State     string
	IpAddress string
	Ssid      string
	Connected bool
}

func NewWpaCliRunner(device string, wpaSocket string) (wcr *WpaCliRunner, err error) {
	wcr = &WpaCliRunner{
		Device:    device,
		WpaSocket: wpaSocket,
	}

	return
}

func (wcr *WpaCliRunner) Check() (state *CurrentWifiState, err error) {
	argv := []string{
		"wpa_cli", "-p", wcr.WpaSocket, "-i", wcr.Device, "status",
	}

	if false {
		log.Printf("Executing %s", strings.Join(argv, " "))
	}

	c := exec.Command(argv[0], argv[1:]...)
	bytes, err := c.Output()
	if err != nil {
		return nil, fmt.Errorf("Command '%s' failed: %v", strings.Join(argv, " "), err)
	}

	linesRe := regexp.MustCompile("(\\S+)=(\\S+)")
	matches := linesRe.FindAllStringSubmatch(string(bytes), -1)
	if matches == nil {
		return nil, nil
	}

	info := make(map[string]string)
	info["wpa_state"] = ""
	info["ip_address"] = ""
	info["ssid"] = ""

	for _, match := range matches {
		info[match[1]] = match[2]
	}

	state = &CurrentWifiState{
		Ssid:      info["ssid"],
		State:     info["wpa_state"],
		IpAddress: info["ip_address"],
		Connected: info["wpa_state"] == "COMPLETED",
	}

	return
}
