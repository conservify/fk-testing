package main

import (
	"log"
	"strings"
)

type TcpDumpRunner struct {
	*BackgroundProcess
	Device string
}

func (tdr *TcpDumpRunner) Line(text string) {

}

func NewTcpDumpRunner(device string) (tdr *TcpDumpRunner, err error) {
	tdr = &TcpDumpRunner{
		Device: device,
	}

	argv := []string{"tcpdump", "-i", device, "-n", "-l"}

	log.Printf("Executing %s", strings.Join(argv, " "))

	bp, err := NewBackgroundProcess("TCP  | ", argv, true, tdr)
	if err != nil {
		return
	}

	tdr.BackgroundProcess = bp

	return
}
