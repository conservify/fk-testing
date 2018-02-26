package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	fkc "github.com/fieldkit/app-protocol/fkdevice"
	fktestutils "github.com/fieldkit/testing/utilities"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type FileOwner struct {
	Uid int
	Gid int
}

func LookupFileOwner() *FileOwner {
	uid, _ := strconv.Atoi(os.Getenv("SUDO_UID"))
	gid, _ := strconv.Atoi(os.Getenv("SUDO_GID"))
	if uid == 0 || gid == 0 {
		return nil
	}

	log.Printf("Creating files as %d:%d", uid, gid)
	return &FileOwner{
		Uid: uid,
		Gid: gid,
	}
}

type options struct {
	Device        string
	DataDirectory string
	WpaSocket     string
	Network       string
	DeviceAddress string

	StartWpa bool

	UploadHost string
	UploadData bool

	fileOwner *FileOwner
}

type StatusEvent struct {
	SSID      string
	Connected bool
	Bound     bool
}

func ConnectAndDownload(ip string, o *options) error {
	dc := &fkc.DeviceClient{
		Address: ip,
		Port:    54321,
	}

	caps, err := dc.QueryCapabilities()
	if err != nil {
		return fmt.Errorf("Unable to get capabilities")
	}

	deviceId := hex.EncodeToString(caps.Capabilities.DeviceId)

	files, err := fktestutils.DownloadDeviceFiles(o.DataDirectory, deviceId, dc)
	if err != nil {
		return fmt.Errorf("Unable to download capabilities")
	}

	if o.fileOwner != nil {
		for _, file := range files {
			os.Chown(file, o.fileOwner.Uid, o.fileOwner.Gid)
		}
	}

	if o.UploadData {
		for _, file := range files {
			if strings.Contains(file, "DATA.BIN") {
				log.Printf("Uploading %s...", file)
				writer := fktestutils.NewStreamingWriter(o.UploadHost)
				df := &fktestutils.DataFile{
					Path: file,
				}
				df.ReadData(writer)
				writer.Finished()
				log.Printf("Done!")
			}
		}
	}

	return nil
}

func main() {
	o := options{}

	flag.StringVar(&o.Device, "device", "", "device to use")
	flag.StringVar(&o.DataDirectory, "data-directory", "./data", "data directory to use")
	flag.StringVar(&o.WpaSocket, "wpa-socket", "", "wpa socket to use")
	flag.StringVar(&o.Network, "network", "", "network")
	flag.StringVar(&o.DeviceAddress, "device-address", "192.168.1.1", "network")

	flag.BoolVar(&o.StartWpa, "start-wpa", false, "start wpa ourselves")

	flag.StringVar(&o.UploadHost, "upload-host", "api.fkdev.org", "host to upload to")
	flag.BoolVar(&o.UploadData, "upload-data", false, "upload data files after downloading")

	flag.Parse()

	o.fileOwner = LookupFileOwner()

	if o.Device == "" {
		os.Exit(2)
	}

	// iw dev
	// ip link show wlx74da387d302a
	// ip link set wlx74da387d302a up
	// iw wlx74da387d302a link

	log.Printf("Starting...")

	scan, err := NewWifiScan(o.Device)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	scan.AddNetwork(o.Network, "", 10)

	networks := scan.ConfiguredNetworks()
	if len(networks) > 0 {
		c := make(chan os.Signal, 1)

		var wg sync.WaitGroup

		wg.Add(1)

		signal.Notify(c, syscall.SIGINT)
		go func() {
			for sig := range c {
				fmt.Printf("\nSignal! %s\n", sig)
				wg.Done()
			}
		}()

		waiting := make([]Waitable, 0)

		if o.StartWpa {
			wsr, err := NewWpaSupplicantRunner(o.Device, networks)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}

			dcr, err := NewDhcpClientRunner(o.Device)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}

			wsr.Start()

			dcr.Start()

			waiting = append(waiting, wsr)
			waiting = append(waiting, dcr)

		}

		wcr, err := NewWpaCliRunner(o.Device, o.WpaSocket)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		log.Printf("Waiting")

		go func() {
			connected := false

			for {
				time.Sleep(1 * time.Second)

				state, err := wcr.Check()
				if err != nil {
					log.Printf("%v", err)
				}

				if state != nil && state.Connected != connected {
					if state.Connected {
						for retries := 3; retries >= 0; retries-- {
							time.Sleep(2 * time.Second)

							err = ConnectAndDownload(o.DeviceAddress, &o)
							if err != nil {
								log.Printf("Error connecting and downloading: %v", err)
							} else {
								break
							}
						}
					}

					connected = state.Connected

				}
			}
		}()

		if !o.StartWpa {
			wg.Wait()
		}

		for _, w := range waiting {
			w.Wait()
		}

		log.Printf("Done")

	}
}
