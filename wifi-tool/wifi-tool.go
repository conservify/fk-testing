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
	"syscall"
	"time"
)

type options struct {
	Device        string
	DataDirectory string
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
		return err
	}

	deviceId := hex.EncodeToString(caps.Capabilities.DeviceId)

	return fktestutils.DownloadDeviceFiles(o.DataDirectory, deviceId, dc)
}

func main() {
	o := options{}

	flag.StringVar(&o.Device, "device", "", "device to use")
	flag.StringVar(&o.DataDirectory, "data-directory", "./data", "data directory to use")

	flag.Parse()

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

	scan.AddNetwork("FK-HELP", "", 10)

	networks := scan.ConfiguredNetworks()
	if len(networks) > 0 {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		go func() {
			for sig := range c {
				fmt.Printf("\nSignal! %s\n", sig)
			}
		}()

		wsr, err := NewWpaSupplicantRunner(o.Device, networks)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		dcr, err := NewDhcpClientRunner(o.Device)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		wcr, err := NewWpaCliRunner(o.Device)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		wsr.Start()

		dcr.Start()

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

							err = ConnectAndDownload("192.168.1.1", &o)
							if err != nil {
								log.Printf("%v", err)
							} else {
								break
							}
						}
					}

					connected = state.Connected

				}
			}
		}()

		wsr.Wait()

		dcr.Wait()

		log.Printf("Done")

	}
}
