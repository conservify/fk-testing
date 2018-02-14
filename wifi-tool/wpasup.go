package main

import (
	"bufio"
	"fmt"
	_ "log"
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

	line := []string{"wpa_supplicant", "-D", "wext", "-i", device, "-c", "./wpa_supplicant.conf"}
	bp, err := NewBackgroundProcess("WPA  | ", line, wsr)
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
	Device string
}

type CurrentWifiState struct {
	State     string
	IpAddress string
	Ssid      string
	Connected bool
}

func NewWpaCliRunner(device string) (wcr *WpaCliRunner, err error) {
	wcr = &WpaCliRunner{
		Device: device,
	}

	return
}

func (wcr *WpaCliRunner) Check() (state *CurrentWifiState, err error) {
	c := exec.Command("wpa_cli", "-p", "/run/wpa_supplicant_fk", "-i", wcr.Device, "status")
	bytes, err := c.Output()
	if err != nil {
		return nil, err
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
