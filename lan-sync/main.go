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
	"time"
)

type DiscoveredDevice struct {
	Address  *net.UDPAddr
	DeviceId []byte
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
			}

			d <- discovery
		}
	}
}

type Device struct {
	Id uint32
}

type Devices struct {
	ById          map[string]*Device
	LastQueryTime map[string]time.Time
}

func (d *Devices) shouldQuery(ip string) bool {
	if !d.LastQueryTime[ip].IsZero() {
		now := time.Now()
		elapsed := now.Sub(d.LastQueryTime[ip])
		if elapsed.Minutes() < 5 {
			return false
		}
	}

	now := time.Now()
	d.LastQueryTime[ip] = now

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

			log.Printf("[%s] Downloading %v", deviceId, file)

			bar := pb.New(int(file.Size)).SetUnits(pb.U_BYTES)
			bar.Start()

			writer := io.MultiWriter(f, bar)

			_, err = dc.DownloadFileToFile(file.Id, 0, writer, nil)
			bar.Set(int(file.Size))
			bar.Finish()

			if err != nil {
				log.Printf("Error: %v", err)
			} else {
				log.Printf("[%s] Erasing %d", deviceId, file.Id)
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

	discoveries := make(chan *DiscoveredDevice)

	devices := &Devices{
		ById:          make(map[string]*Device),
		LastQueryTime: make(map[string]time.Time),
	}

	go discoverDevicesOnLocalNetwork(discoveries)

	for discovered := range discoveries {
		ip := discovered.Address.IP.String()
		if devices.shouldQuery(ip) {
			log.Printf("Querying %v...", discovered)

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
				log.Printf("Bad reply")
				continue
			}

			deviceId := hex.EncodeToString(reply.Capabilities.DeviceId)
			devices.addDevice(deviceId)
			downloadDeviceFiles(deviceId, dc)
		} else {
			log.Printf("Seen %v", discovered)
		}
	}
}
