package main

import (
	_ "bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

type DeviceLocation struct {
	Time        uint32
	Coordinates [3]float32
}

type SensorReading struct {
	Time   uint32
	Value  float32
	Status uint32
}

type DataEntry struct {
	Location DeviceLocation
	Reading  SensorReading
}

type DataFile struct {
	Path string
}

func (df *DataFile) ReadData(dw DataWriter) {
	file, err := os.Open(df.Path)
	if err != nil {
		log.Fatal("Error while opening file", err)
	}

	defer file.Close()

	for {
		entry := DataEntry{}
		err = binary.Read(file, binary.LittleEndian, &entry)
		if err == io.EOF {
			return
		}
		dw.Write(df, &entry)
	}
}

type DataWriter interface {
	Write(df *DataFile, entry *DataEntry)
}

type LogDataWriter struct {
}

func (ldw *LogDataWriter) Write(df *DataFile, entry *DataEntry) {
	log.Printf("%s: %+v", df.Path, entry)

}

type CsvDataWriter struct {
}

func (csv *CsvDataWriter) Write(df *DataFile, entry *DataEntry) {
	fmt.Printf("%s,%d,%f,%f,%f,%d,%f\n", df.Path, entry.Location.Time, entry.Location.Coordinates[0], entry.Location.Coordinates[1], entry.Location.Coordinates[2], entry.Reading.Time, entry.Reading.Value)
}

type options struct {
	Csv bool
}

func main() {
	o := options{}

	flag.BoolVar(&o.Csv, "csv", false, "write csv")

	flag.Parse()

	var writer DataWriter

	if o.Csv {
		writer = &CsvDataWriter{}
	} else {
		writer = &LogDataWriter{}
	}

	for _, path := range flag.Args() {
		log.Printf("Opening %s", path)

		df := &DataFile{
			Path: path,
		}

		df.ReadData(writer)
	}
}
