package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	// fktesting "github.com/fieldkit/cloud/server/api/tool"
	"github.com/fieldkit/cloud/server/backend/ingestion"
	pb "github.com/fieldkit/data-protocol"
	"github.com/golang/protobuf/proto"
	"io"
	"io/ioutil"
	"log"
)

type DataFile struct {
	Path string
}

func (df *DataFile) ReadData(dw DataWriter) {
	data, err := ioutil.ReadFile(df.Path)
	if err != nil {
		log.Fatal("Error while opening file", err)
	}

	buf := proto.NewBuffer(data[:])

	for {
		messageBytes, err := buf.DecodeRawBytes(true)
		if err != nil {
			if err == io.ErrUnexpectedEOF {
				err = nil
				break
			}
			log.Printf("%v", err)
			return
		}

		messageBuffer := proto.NewBuffer(messageBytes[:])
		record := new(pb.DataRecord)
		err = messageBuffer.Unmarshal(record)
		if err != nil {
			log.Printf("Error: %v", err)
			log.Printf("Length: %v", len(messageBuffer.Bytes()))
			return
		}

		dw.Write(df, record)
	}
}

type DataWriter interface {
	Write(df *DataFile, record *pb.DataRecord)
}

type LogDataWriter struct {
	options         *options
	deviceId        string
	location        *pb.DeviceLocation
	time            int64
	numberOfSensors uint32
	readingsSeen    uint32
	sensors         map[uint32]*pb.SensorInfo
	readings        map[uint32]float32
}

func (ldw *LogDataWriter) CreateFieldKitMessage() *ingestion.HttpJsonMessage {
	values := make(map[string]string)
	for key, value := range ldw.readings {
		values[ldw.sensors[key].Name] = fmt.Sprintf("%f", value)
	}

	return &ingestion.HttpJsonMessage{
		Location: []float64{float64(ldw.location.Longitude), float64(ldw.location.Latitude), float64(ldw.location.Altitude)},
		Time:     ldw.time,
		Device:   ldw.deviceId,
		Stream:   "",
		Values:   values,
	}
}

func mapOfFloatsToMapOfStrings(original map[string]float32) map[string]string {
	r := make(map[string]string)
	for key, value := range original {
		r[key] = fmt.Sprintf("%f", value)
	}
	return r
}

func (ldw *LogDataWriter) Write(df *DataFile, record *pb.DataRecord) {
	if record.Metadata != nil {
		if ldw.deviceId == "" {
			ldw.deviceId = hex.EncodeToString(record.Metadata.DeviceId)
		}
		if record.Metadata.Sensors != nil {
			if ldw.numberOfSensors == 0 {
				for _, sensor := range record.Metadata.Sensors {
					ldw.sensors[sensor.Sensor] = sensor
					ldw.numberOfSensors += 1
				}
				log.Printf("Found %d sensors", ldw.numberOfSensors)
			}
		}

	}
	if record.LoggedReading != nil {
		if record.LoggedReading.Location != nil {
			ldw.location = record.LoggedReading.Location
		}
		reading := record.LoggedReading.Reading
		if reading != nil {
			ldw.readings[reading.Sensor] = reading.Value
			ldw.readingsSeen += 1

			if ldw.readingsSeen == ldw.numberOfSensors {
				ldw.time = int64(record.LoggedReading.Reading.Time)

				if ldw.location != nil {
					b, err := json.Marshal(ldw.CreateFieldKitMessage())
					if err != nil {
						log.Fatalf("Error %v", err)
					}

					if true {
						body := bytes.NewBufferString(string(b))
						url := fmt.Sprintf("%s://%s/messages/ingestion", ldw.options.Scheme, ldw.options.Host)
						url += "?token=" + "IGNORED"
						_, err = http.Post(url, ingestion.HttpProviderJsonContentType, body)
						if err != nil {
							log.Fatalf("%s %s", url, err)
						}
					}

					fmt.Println(string(b))
				}

				ldw.readingsSeen = 0
			}
		}
	}

}

type CsvDataWriter struct {
}

func (csv *CsvDataWriter) Write(df *DataFile, record *pb.DataRecord) {
	entry := record.LoggedReading
	fmt.Printf("%s,%d,%f,%f,%f,%d,%f\n", df.Path, entry.Location.Time, entry.Location.Longitude, entry.Location.Latitude, entry.Location.Altitude, entry.Reading.Time, entry.Reading.Value)
}

type options struct {
	Csv bool

	Project    string
	Expedition string
	DeviceName string
	Host       string
	Scheme     string
	Username   string
	Password   string
}

func main() {
	o := options{}

	flag.BoolVar(&o.Csv, "csv", false, "write csv")

	flag.StringVar(&o.Project, "project", "www", "project")
	flag.StringVar(&o.DeviceName, "device-name", "weather-proxy", "device name")
	flag.StringVar(&o.Scheme, "scheme", "http", "scheme to use")
	flag.StringVar(&o.Host, "host", "127.0.0.1:8080", "hostname to use")
	flag.StringVar(&o.Username, "username", "demo-user", "username to use")
	flag.StringVar(&o.Password, "password", "asdfasdfasdf", "password to use")

	flag.Parse()

	var writer DataWriter

	if o.Csv {
		writer = &CsvDataWriter{}
	} else {
		writer = &LogDataWriter{
			options:  &o,
			sensors:  make(map[uint32]*pb.SensorInfo),
			readings: make(map[uint32]float32),
		}
	}

	for _, path := range flag.Args() {
		log.Printf("Opening %s", path)

		df := &DataFile{
			Path: path,
		}

		df.ReadData(writer)
	}
}
