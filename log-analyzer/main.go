package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/hpcloud/tail"
	_ "log"
	"os"
	"regexp"
	"strconv"
	"time"
)

type DeviceEvent struct {
}

type DeviceRun struct {
	RestartReason string
	Uptime        uint32
	Events        []DeviceEvent
}

type DeviceState struct {
	Runs []DeviceRun
}

type LogFileParser struct {
}

type LogFileEntry struct {
	Uptime   uint32
	Facility string
	Message  string
}

var simpleLineRegex = regexp.MustCompile("(\\d+)\\s+(\\S+)\\s+(.+)")

func (p *LogFileParser) ProcessLine(line string) error {
	v := simpleLineRegex.FindAllStringSubmatch(line, -1)
	if v == nil {
		return nil
	}

	uptime, err := strconv.Atoi(v[0][1])
	if err != nil {
		return fmt.Errorf("Unable to parse uptime: %v", err)
	}

	lfe := LogFileEntry{
		Uptime:   uint32(uptime),
		Facility: v[0][2],
		Message:  v[0][3],
	}

	p.ProcessEntry(&lfe)

	return nil
}

type StatusUpdate struct {
	Battery float32
	Memory  uint32
	IP      string
}

type SensorReading struct {
	Time       uint32
	SensorName string
	SensorId   int32
	Value      float64
}

type Location struct {
	Time       uint32
	Satellites uint32
	HDOP       uint32
	Longitude  float64
	Latitude   float64
	Altitude   float64
}

func (p *LogFileParser) ProcessEntry(lfe *LogFileEntry) error {
	schedulerRegex := regexp.MustCompile("(.+): run task \\(again = .+\\)")
	v := schedulerRegex.FindAllStringSubmatch(lfe.Message, -1)
	if v != nil {
		parsed, err := time.Parse("2006/01/02 15:04:05", v[0][1])
		if err != nil {
			return fmt.Errorf("Unable to parse time: %v", err)
		}
		fmt.Printf("Scheduler: %v\n", parsed)
	}

	gpsRegex := regexp.MustCompile("Time\\((\\d+)\\) Sats\\((\\d+)\\) Hdop\\((\\d+)\\) Loc\\(([-\\d.]+), ([-\\d.]+)\\, ([-\\d.]+)\\)")
	v = gpsRegex.FindAllStringSubmatch(lfe.Message, -1)
	if v != nil {
		time, err := strconv.Atoi(v[0][1])
		if err != nil {
			return fmt.Errorf("Unable to parse: %v", err)
		}
		satellites, err := strconv.Atoi(v[0][2])
		if err != nil {
			return fmt.Errorf("Unable to parse: %v", err)
		}
		hdop, err := strconv.Atoi(v[0][3])
		if err != nil {
			return fmt.Errorf("Unable to parse: %v", err)
		}
		longitude, err := strconv.ParseFloat(v[0][4], 32)
		if err != nil {
			return fmt.Errorf("Unable to parse: %v", err)
		}
		latitude, err := strconv.ParseFloat(v[0][5], 32)
		if err != nil {
			return fmt.Errorf("Unable to parse: %v", err)
		}
		altitude, err := strconv.ParseFloat(v[0][6], 32)
		if err != nil {
			return fmt.Errorf("Unable to parse: %v", err)
		}
		location := Location{
			Time:       uint32(time),
			Satellites: uint32(satellites),
			HDOP:       uint32(hdop),
			Longitude:  longitude,
			Latitude:   latitude,
			Altitude:   altitude,
		}
		fmt.Printf("GPS: %v\n", location)
	}

	readingRegex := regexp.MustCompile("Appended reading \\(\\d+ bytes\\) \\((\\d+), (\\d+), '(.+)' = (\\S+)\\)")
	v = readingRegex.FindAllStringSubmatch(lfe.Message, -1)
	if v != nil {
		time, err := strconv.Atoi(v[0][1])
		if err != nil {
			return fmt.Errorf("Unable to parse reading time: %v", err)
		}
		sensorId, err := strconv.Atoi(v[0][2])
		if err != nil {
			return fmt.Errorf("Unable to parse reading value: %v", err)
		}
		sensorName := v[0][3]
		value, err := strconv.ParseFloat(v[0][4], 32)
		if err != nil {
			return fmt.Errorf("Unable to parse reading value: %v", err)
		}

		reading := SensorReading{
			Time:       uint32(time),
			SensorId:   int32(sensorId),
			SensorName: sensorName,
			Value:      value,
		}

		fmt.Printf("Reading: %v\n", reading)
	}

	statusRegex := regexp.MustCompile("Status \\(([\\d.]+)% / ([\\d.]+)mv\\) \\((\\d+) free\\) \\((\\S+)\\)")
	v = statusRegex.FindAllStringSubmatch(lfe.Message, -1)
	if v != nil {
		battery, err := strconv.ParseFloat(v[0][1], 32)
		if err != nil {
			return fmt.Errorf("Unable to parse battery: %v", err)
		}
		memory, err := strconv.Atoi(v[0][3])
		if err != nil {
			return fmt.Errorf("Unable to parse memory: %v", err)
		}
		status := StatusUpdate{
			Battery: float32(battery),
			Memory:  uint32(memory),
			IP:      v[0][4],
		}
		fmt.Printf("Status: %v\n", status)
	}

	resetCauseRegex := regexp.MustCompile("ResetCause: (.+)")
	v = resetCauseRegex.FindAllStringSubmatch(lfe.Message, -1)
	if v != nil {
		fmt.Printf("ResetCause: %v\n", v[0][1])
	}

	longTaskRegex := regexp.MustCompile("Long tick from (\\S+) \\((\\d+)\\)")
	v = longTaskRegex.FindAllStringSubmatch(lfe.Message, -1)
	if v != nil {
		fmt.Printf("Long Tick: %v %v %v\n", lfe.Facility, v[0][1], v[0][2])
	}

	return nil
}

type options struct {
	Follow bool
}

func main() {
	o := options{}

	flag.BoolVar(&o.Follow, "follow", false, "follow the file")

	flag.Parse()

	parser := &LogFileParser{}
	args := flag.Args()
	if len(args) > 0 {
		for {
			t, err := tail.TailFile(args[0], tail.Config{Follow: o.Follow})
			if err == nil {
				for line := range t.Lines {
					parser.ProcessLine(line.Text)
				}
			}

			if o.Follow {
				time.Sleep(1 * time.Second)
			} else {
				break
			}
		}
	} else {
		r := bufio.NewScanner(os.Stdin)
		for r.Scan() {
			l := r.Text()
			parser.ProcessLine(l)
		}
	}
}
