package utilities

import (
	"fmt"
	pb "github.com/fieldkit/data-protocol"
)

type CsvDataWriter struct {
}

func (csv *CsvDataWriter) Write(df *DataFile, record *pb.DataRecord, raw []byte) error {
	if record.LoggedReading != nil && record.LoggedReading.Location != nil && record.LoggedReading.Reading != nil {
		entry := record.LoggedReading
		fmt.Printf("%s,%d,%f,%f,%f,%d,%d,%f\n", df.Path, entry.Location.Time, entry.Location.Longitude, entry.Location.Latitude, entry.Location.Altitude, entry.Reading.Sensor, entry.Reading.Time, entry.Reading.Value)
	}
	return nil
}

func (csv *CsvDataWriter) Finished() error {
	return nil
}
