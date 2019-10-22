package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"os/exec"
	"sync"
	"time"
	// fktestutils "github.com/conservify/fk-testing/utilities"
	// fkc "github.com/fieldkit/app-protocol/fkdevice"
)

type DiscoveredDevice struct {
	Address  *net.UDPAddr
	DeviceId []byte
	Time     time.Time
}

type Device struct {
	Id             uint32
	Busy           bool
	LastQueryTime  time.Time
	LastNotifyTime time.Time
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

func secondsSince(t time.Time) float64 {
	now := time.Now()
	elapsed := now.Sub(t)
	return elapsed.Seconds()
}

func (d *Devices) shouldQuery(id string, interval float64) bool {
	if d.ById[id] == nil {
		return true
	}

	device := d.ById[id]
	if !device.LastQueryTime.IsZero() {
		if secondsSince(device.LastQueryTime) < interval {
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

func (d *Devices) shouldNotify(id string) bool {
	if d.ById[id] == nil {
		return true
	}

	device := d.ById[id]
	if !device.LastNotifyTime.IsZero() {
		if secondsSince(device.LastNotifyTime) < 3*60 {
			return false
		}
	}

	now := time.Now()
	device.LastNotifyTime = now

	return true
}

func (d *Devices) addDevice(id string) {
	if d.ById[id] == nil {
		d.ById[id] = &Device{}
	}
}

func (d *Devices) query(ip, deviceId string, o *options) {
	/*
		dc := &fkc.DeviceClient{
			Address: ip,
			Port:    54321,
		}

			capabilitiesReply, err := dc.QueryCapabilities()
			if err != nil {
				log.Printf("Error: %v", err)
				return
			}
			if capabilitiesReply == nil || capabilitiesReply.Capabilities == nil {
				log.Printf("Error: Bad reply")
				return
			}

			statusReply, err := dc.QueryStatus()
			if err != nil {
				log.Printf("Error: %v", err)
				return
			}
			if statusReply == nil || statusReply.Status == nil {
				log.Printf("Error: Bad reply")
				return
			}

			log.Printf("%s: %v", ip, statusReply.Status)

			d.addDevice(deviceId)
			d.markBusy(deviceId)

			if o.Download {
				fktestutils.DownloadDeviceFiles("data", deviceId, dc)
			}
			d.markAvailable(deviceId)
	*/
}

type options struct {
	Query    bool
	Download bool
	Notify   bool
}

func notify(title, text string) {
	cmd := exec.Command("notify-send", title, text)
	cmd.Run()
}

func main() {
	o := &options{}

	flag.BoolVar(&o.Query, "query", false, "query devices")
	flag.BoolVar(&o.Download, "download", false, "downlaod files from devices")
	flag.BoolVar(&o.Notify, "notify", false, "Use notify-send")

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
		age := secondsSince(discovered.Time)

		if age < 60 {
			if o.Query && devices.shouldQuery(deviceId, 60) {
				devices.query(ip, deviceId, o)
			} else {
				log.Printf("%v %v", ip, deviceId)
			}

			if o.Notify && devices.shouldNotify(deviceId) {
				notify("fk-lan-sync", fmt.Sprintf("Device: %s", ip))
			}
		}
	}
}
