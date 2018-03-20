package utilities

import (
	"fmt"

	pb "github.com/fieldkit/data-protocol"
)

type CsvDataWriter struct {
}

func (csv *CsvDataWriter) Begin(df *DataFile, chain BeginChainFunc) error {
	return chain(df)
}

func (csv *CsvDataWriter) Process(df *DataFile, record *pb.DataRecord, begin BeginChainFunc, chain ProcessChainFunc, end EndChainFunc) error {
	if record.LoggedReading != nil && record.LoggedReading.Location != nil && record.LoggedReading.Reading != nil {
		entry := record.LoggedReading
		fmt.Printf("%s,%d,%f,%f,%f,%d,%d,%f\n", df.Path, entry.Location.Time, entry.Location.Longitude, entry.Location.Latitude, entry.Location.Altitude, entry.Reading.Sensor, entry.Reading.Time, entry.Reading.Value)
	}
	return chain(df, record)
}

func (csv *CsvDataWriter) End(df *DataFile, chain EndChainFunc) error {
	return chain(df)
}
