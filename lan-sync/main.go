package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	fkc "github.com/fieldkit/app-protocol/fkdevice"
	fktestutils "github.com/fieldkit/testing/utilities"
)

type DiscoveredDevice struct {
	Address  *net.UDPAddr
	DeviceId []byte
	Time     time.Time
}

type Device struct {
	Id            uint32
	Busy          bool
	LastQueryTime time.Time
}

type Devices struct {
	Lock *sync.Mutex
	ById map[string]*Device
}

func (d *Devices) markBusy(id string) {
	d.ById[id].Busy = true
}

func (d *Devices) markAvailable(id string) {
	d.ById[id].Busy = false
}

func discoverDevicesOnLocalNetwork(d chan *DiscoveredDevice) {
	for {
		serverAddr, err := net.ResolveUDPAddr("udp", ":54321")
		if err != nil {
			fmt.Println("Error: ", err)
		}

		serverConn, err := net.ListenUDP("udp", serverAddr)
		if err != nil {
			fmt.Println("Error: ", err)
		}

		defer serverConn.Close()

		for {
			buf := make([]byte, 32)
			len, addr, err := serverConn.ReadFromUDP(buf)
			if err != nil {
				fmt.Println("Error: ", err)
			}

			// log.Printf("DeviceId: %v %d bytes %s", addr.IP.String(), len, hex.EncodeToString(buf[0:len]))

			discovery := &DiscoveredDevice{
				Address:  addr,
				DeviceId: buf[0:len],
				Time:     time.Now(),
			}

			d <- discovery
		}
	}
}

func minutesSince(t time.Time) float64 {
	now := time.Now()
	elapsed := now.Sub(t)
	return elapsed.Minutes()
}

func (d *Devices) shouldQuery(id string) bool {
	if d.ById[id] == nil {
		return true
	}

	device := d.ById[id]
	if !device.LastQueryTime.IsZero() {
		if minutesSince(device.LastQueryTime) < 5 {
			return false
		}
	}

	if !device.Busy {
		now := time.Now()
		device.LastQueryTime = now
		return true
	}

	return true
}

func (d *Devices) addDevice(id string) {
	if d.ById[id] == nil {
		d.ById[id] = &Device{}
	}
}

type options struct {
	Download bool
}

func main() {
	o := options{}

	flag.BoolVar(&o.Download, "download", false, "downlaod files from devices")

	flag.Parse()

	log.Printf("Starting...")

	discoveries := make(chan *DiscoveredDevice)

	devices := &Devices{
		Lock: &sync.Mutex{},
		ById: make(map[string]*Device),
	}

	go discoverDevicesOnLocalNetwork(discoveries)

	for discovered := range discoveries {
		ip := discovered.Address.IP.String()
		deviceId := hex.EncodeToString(discovered.DeviceId)
		age := minutesSince(discovered.Time)

		if age < 1 {
			if devices.shouldQuery(deviceId) {
				dc := &fkc.DeviceClient{
					Address: ip,
					Port:    54321,
				}

				capabilitiesReply, err := dc.QueryCapabilities()
				if err != nil {
					log.Printf("Error: %v", err)
					continue
				}
				if capabilitiesReply == nil || capabilitiesReply.Capabilities == nil {
					log.Printf("Error: Bad reply")
					continue
				}

				statusReply, err := dc.QueryStatus()
				if err != nil {
					log.Printf("Error: %v", err)
					continue
				}
				if statusReply == nil || statusReply.Status == nil {
					log.Printf("Error: Bad reply")
					continue
				}

				log.Printf("Status: %v", statusReply.Status)

				devices.addDevice(deviceId)
				devices.markBusy(deviceId)

				if o.Download {
					fktestutils.DownloadDeviceFiles("data", deviceId, dc)
				}
				devices.markAvailable(deviceId)
			} else {
				log.Printf("%v %v", discovered.Address.IP.String(), deviceId)
			}
		}
	}
}
