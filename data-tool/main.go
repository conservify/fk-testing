package main

import (
	"flag"
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

		dw.Write(df, record, messageBytes)
	}
}

type DataWriter interface {
	Write(df *DataFile, record *pb.DataRecord, raw []byte) error
	Finished() error
}

type NullWriter struct {
}

func (w *NullWriter) Write(df *DataFile, record *pb.DataRecord, raw []byte) error {
	return nil
}

func (w *NullWriter) Finished() error {
	return nil
}

type options struct {
	Csv        bool
	PostJson   bool
	PostStream bool

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
	flag.BoolVar(&o.PostJson, "post-json", false, "interpret and post json")
	flag.BoolVar(&o.PostStream, "post-stream", false, "post binary stream directly")

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
	} else if o.PostStream {
		writer = &StreamingWriter{
			options: &o,
		}
	} else if o.PostJson {
		writer = &DataBinaryToPostWriter{
			options:  &o,
			sensors:  make(map[uint32]*pb.SensorInfo),
			readings: make(map[uint32]float32),
		}
	} else {
		writer = &NullWriter{}
	}

	log.Printf("Using %T", writer)

	for _, path := range flag.Args() {
		log.Printf("Opening %s", path)

		df := &DataFile{
			Path: path,
		}

		df.ReadData(writer)
	}

	writer.Finished()
}
