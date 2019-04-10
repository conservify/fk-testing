package main

import (
	_ "log"
	"regexp"
)

type DhcpClientRunner struct {
	BackgroundProcessCallback
	*BackgroundProcess
	Bound bool
}

func (dcr *DhcpClientRunner) Line(text string) {
	boundRe := regexp.MustCompile("bound to (\\S+)")
	boundM := boundRe.FindStringSubmatch(text)
	if boundM != nil {
		dcr.Bound = true
	}
}

func NewDhcpClientRunner(device string) (dcr *DhcpClientRunner, err error) {
	dcr = &DhcpClientRunner{}

	line := []string{"dhclient", "-d", device, "-sf", "./dhclient-script", "-cf", "dhclient.conf"}
	bp, err := NewBackgroundProcess("DHCP | ", line, false, dcr)
	if err != nil {
		return
	}

	dcr.BackgroundProcess = bp

	return
}
