package utilities

import (
	"time"
	"fmt"

	pb "github.com/fieldkit/data-protocol"
)

type CsvDataWriter struct {
	FormattedTimes bool
}

func (csv *CsvDataWriter) Begin(df *DataFile, chain BeginChainFunc) error {
	return chain(df)
}

func (csv *CsvDataWriter) Process(df *DataFile, record *pb.DataRecord, begin BeginChainFunc, chain ProcessChainFunc, end EndChainFunc) error {
	if record.LoggedReading != nil && record.LoggedReading.Location != nil && record.LoggedReading.Reading != nil {
		entry := record.LoggedReading

		formattedLocationTime := csv.formatUnixTime(entry.Location.Time)
		formattedReadingTime := csv.formatUnixTime(entry.Reading.Time)

		fmt.Printf("%s,%s,%f,%f,%f,%d,%s,%f\n", df.Path, formattedLocationTime, entry.Location.Longitude, entry.Location.Latitude, entry.Location.Altitude, entry.Reading.Sensor, formattedReadingTime, entry.Reading.Value)
	}
	return chain(df, record)
}

func (csv *CsvDataWriter) End(df *DataFile, chain EndChainFunc) error {
	return chain(df)
}

func (csv *CsvDataWriter) formatUnixTime(unix uint64) string {
	if csv.FormattedTimes {
		t := time.Unix(int64(unix), 0)
		return t.Format("01/02/2006 15:04:05 -0700 MST")
	}
	return fmt.Sprintf("%d", unix)
}
