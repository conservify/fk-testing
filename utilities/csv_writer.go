package utilities

import (
	"fmt"
	pb "github.com/fieldkit/data-protocol"
)

type CsvDataWriter struct {
}

func (csv *CsvDataWriter) Write(df *DataFile, record *pb.DataRecord, raw []byte) error {
	entry := record.LoggedReading
	fmt.Printf("%s,%d,%f,%f,%f,%d,%f\n", df.Path, entry.Location.Time, entry.Location.Longitude, entry.Location.Latitude, entry.Location.Altitude, entry.Reading.Time, entry.Reading.Value)
	return nil
}

func (csv *CsvDataWriter) Finished() error {
	return nil
}
