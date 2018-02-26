package main

import (
	"flag"
	fktestutils "github.com/fieldkit/testing/utilities"
	"log"
)

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

	var writer fktestutils.DataWriter

	if o.Csv {
		writer = &fktestutils.CsvDataWriter{}
	} else if o.PostStream {
		writer = fktestutils.NewStreamingWriter(o.Host)
	} else if o.PostJson {
		writer = fktestutils.NewDataBinaryToPostWriter(o.Scheme, o.Host)
	} else {
		writer = &fktestutils.NullWriter{}
	}

	log.Printf("Using %T", writer)

	for _, path := range flag.Args() {
		log.Printf("Opening %s", path)

		df := &fktestutils.DataFile{
			Path: path,
		}

		df.ReadData(writer)
	}

	writer.Finished()
}
