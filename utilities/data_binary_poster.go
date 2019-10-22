package utilities

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/fieldkit/cloud/server/backend/ingestion/formatting"
	pb "github.com/fieldkit/data-protocol"
)

type DataBinaryToPostWriter struct {
	scheme          string
	host            string
	deviceId        string
	location        *pb.DeviceLocation
	time            int64
	numberOfSensors uint32
	readingsSeen    uint32
	sensors         map[uint32]*pb.SensorInfo
	readings        map[uint32]float32
}

func NewDataBinaryToPostWriter(scheme, host string) *DataBinaryToPostWriter {
	return &DataBinaryToPostWriter{
		scheme:   scheme,
		host:     host,
		sensors:  make(map[uint32]*pb.SensorInfo),
		readings: make(map[uint32]float32),
	}
}

func (dbpw *DataBinaryToPostWriter) CreateHttpJsonMessage() *formatting.HttpJsonMessage {
	values := make(map[string]interface{})
	for key, value := range dbpw.readings {
		values[dbpw.sensors[key].Name] = fmt.Sprintf("%f", value)
	}

	return &formatting.HttpJsonMessage{
		Location: []float64{float64(dbpw.location.Longitude), float64(dbpw.location.Latitude), float64(dbpw.location.Altitude)},
		Time:     dbpw.time,
		Device:   dbpw.deviceId,
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

func (dbpw *DataBinaryToPostWriter) Begin(df *DataFile, chain BeginChainFunc) error {
	return chain(df)
}

func (dbpw *DataBinaryToPostWriter) Process(df *DataFile, record *pb.DataRecord, begin BeginChainFunc, chain ProcessChainFunc, end EndChainFunc) error {
	if record.Metadata != nil {
		if dbpw.deviceId == "" {
			dbpw.deviceId = hex.EncodeToString(record.Metadata.DeviceId)
		}
		if record.Metadata.Sensors != nil {
			if dbpw.numberOfSensors == 0 {
				for _, sensor := range record.Metadata.Sensors {
					dbpw.sensors[sensor.Number] = sensor
					dbpw.numberOfSensors += 1
				}
				log.Printf("Found %d sensors", dbpw.numberOfSensors)
			}
		}

	}
	if record.LoggedReading != nil && dbpw.numberOfSensors > 0 {
		if record.LoggedReading.Location != nil {
			dbpw.location = record.LoggedReading.Location
		}
		reading := record.LoggedReading.Reading
		if reading != nil {
			if record.LoggedReading.Location == nil || record.LoggedReading.Location.Fix != 1 {
				log.Printf("Skip unfixed reading")
				return nil
			}
			dbpw.readings[reading.Sensor] = reading.Value
			dbpw.readingsSeen += 1

			if dbpw.readingsSeen == dbpw.numberOfSensors {
				dbpw.time = int64(record.LoggedReading.Reading.Time)

				if dbpw.location != nil {
					b, err := json.Marshal(dbpw.CreateHttpJsonMessage())
					if err != nil {
						log.Fatalf("Error %v", err)
					}

					if false {
						body := bytes.NewBufferString(string(b))
						url := fmt.Sprintf("%s://%s/messages/ingestion", dbpw.scheme, dbpw.host)
						url += "?token=" + "IGNORED"
						_, err = http.Post(url, formatting.HttpProviderJsonContentType, body)
						if err != nil {
							log.Fatalf("%s %s", url, err)
						}

						fmt.Println(string(b))
					}
				}

				dbpw.readingsSeen = 0
			}
		}
	}
	return chain(df, record)

}

func (dbpw *DataBinaryToPostWriter) End(df *DataFile, chain EndChainFunc) error {
	return chain(df)
}
