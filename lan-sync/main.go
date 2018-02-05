package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	fkc "github.com/fieldkit/app-protocol/device"
	"gopkg.in/cheggaaa/pb.v1"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
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

		buf := make([]byte, 1024)

		for {
			len, addr, err := serverConn.ReadFromUDP(buf)
			if err != nil {
				fmt.Println("Error: ", err)
			}

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

func downloadDeviceFiles(deviceId string, dc *fkc.DeviceClient) {
	fileReply, err := dc.QueryFiles()
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	if fileReply == nil || fileReply.Files == nil || fileReply.Files.Files == nil {
		log.Printf("Error, no files in reply")
		return
	}

	for _, file := range fileReply.Files.Files {
		if file.Size > 0 {
			dir := fmt.Sprintf("data/%s", deviceId)
			stamp := time.Now().Format("20060102_150405")
			fileName := fmt.Sprintf("%s/%s_%s_%d", dir, file.Name, stamp, file.Version)

			err = os.MkdirAll(dir, 0777)
			if err != nil {
				log.Fatalf("Unable to create %s (%v)", dir, err)
			}
			f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
			if err != nil {
				log.Fatalf("Unable to open %s (%v)", fileName, err)
			}

			defer f.Close()

			log.Printf("[%s] Downloading %v (%s)", deviceId, file, fileName)

			bar := pb.New(int(file.Size)).SetUnits(pb.U_BYTES)
			bar.Start()

			writer := io.MultiWriter(f, bar)

			_, err = dc.DownloadFileToFile(file.Id, 65536*2, writer, nil)

			bar.Set(int(file.Size))
			bar.Finish()

			if err != nil {
				log.Printf("Error: %v", err)
			} else {
				dc.EraseFile(file.Id)
				if err != nil {
					log.Printf("Error: %v", err)
				}
			}
		}
	}
}

func main() {
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

				reply, err := dc.QueryCapabilities()
				if err != nil {
					log.Printf("Error: %v", err)
					continue
				}
				if reply == nil || reply.Capabilities == nil {
					log.Printf("Error: Bad reply")
					continue
				}

				devices.addDevice(deviceId)
				devices.markBusy(deviceId)
				downloadDeviceFiles(deviceId, dc)
				devices.markAvailable(deviceId)
			} else {
				log.Printf("%v %v", discovered.Address.IP.String(), deviceId)
			}
		}
	}
}
