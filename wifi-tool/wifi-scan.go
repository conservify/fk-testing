package main

import (
	"bufio"
	"os/exec"
	"regexp"
	"strings"
)

type WifiNetwork struct {
	Name     string
	Password string
	Priority int
}

type WifiScan struct {
	Device   string
	Networks []*WifiNetwork
}

func NewWifiScan(device string) (scan *WifiScan, err error) {
	scan = &WifiScan{
		Device:   device,
		Networks: make([]*WifiNetwork, 0),
	}

	return
}

func (s *WifiScan) Scan() error {
	s.Networks = make([]*WifiNetwork, 0)
	cmd := exec.Command("iw", s.Device, "scan")
	raw, err := cmd.Output()
	if err != nil {
		return err
	}

	ssidRe := regexp.MustCompile("SSID: (.+)")
	scanner := bufio.NewScanner(strings.NewReader(string(raw)))
	for scanner.Scan() {
		ssidM := ssidRe.FindStringSubmatch(scanner.Text())
		if ssidM != nil {
			s.Networks = append(s.Networks, &WifiNetwork{
				Name:     ssidM[1],
				Priority: -1,
			})
		}
	}

	return nil
}

func (s *WifiScan) AddNetwork(name string, password string, priority int) {
	s.Networks = append(s.Networks, &WifiNetwork{
		Name:     name,
		Password: password,
		Priority: priority,
	})
}

func (s *WifiScan) ConfiguredNetworks() []*WifiNetwork {
	networks := make([]*WifiNetwork, 0)
	for _, n := range s.Networks {
		if n.Priority >= 0 {
			networks = append(networks, n)
		}
	}
	return networks
}

func (s *WifiScan) HasNetwork(pattern string) *WifiNetwork {
	re := regexp.MustCompile(pattern)
	for _, n := range s.Networks {
		if re.MatchString(n.Name) {
			return n
		}
	}
	return nil
}
