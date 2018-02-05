package main

import (
	_ "bytes"
	"flag"
	"fmt"
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
			return
		}

		messageBuffer := proto.NewBuffer(messageBytes[:])
		record := new(pb.DataRecord)
		err = messageBuffer.Unmarshal(record)
		if err != nil {
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
}

func (ldw *LogDataWriter) Write(df *DataFile, record *pb.DataRecord) {
	log.Printf("%s: %+v", df.Path, record)

}

type CsvDataWriter struct {
}

func (csv *CsvDataWriter) Write(df *DataFile, record *pb.DataRecord) {
	entry := record.LoggedReading
	fmt.Printf("%s,%d,%f,%f,%f,%d,%f\n", df.Path, entry.Location.Time, entry.Location.Longitude, entry.Location.Latitude, entry.Location.Altitude, entry.Reading.Time, entry.Reading.Value)
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
