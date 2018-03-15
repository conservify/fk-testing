package main

import (
	"flag"
	"log"

	fktestutils "github.com/fieldkit/testing/utilities"
)

type options struct {
	Csv        bool
	PostJson   bool
	PostStream bool
	Verbose    bool
	Split      int

	Project    string
	Expedition string
	DeviceName string
	Host       string
	Scheme     string
	Username   string
	Password   string

	DeviceId string
}

func main() {
	o := options{}

	flag.BoolVar(&o.Csv, "csv", false, "write csv")
	flag.BoolVar(&o.PostJson, "post-json", false, "interpret and post json")
	flag.BoolVar(&o.PostStream, "post-stream", false, "post binary stream directly")
	flag.BoolVar(&o.Verbose, "verbose", false, "verbose output")

	flag.IntVar(&o.Split, "split", 0, "split stream into smaller files")

	flag.StringVar(&o.Project, "project", "www", "project")
	flag.StringVar(&o.DeviceName, "device-name", "weather-proxy", "device name")
	flag.StringVar(&o.Scheme, "scheme", "http", "scheme to use")
	flag.StringVar(&o.Host, "host", "127.0.0.1:8080", "hostname to use")
	flag.StringVar(&o.Username, "username", "demo-user", "username to use")
	flag.StringVar(&o.Password, "password", "asdfasdfasdf", "password to use")

	flag.StringVar(&o.DeviceId, "force-device-id", "", "force uploaded data to use this device id")

	flag.Parse()

	var writer fktestutils.DataWriter

	if o.Csv {
		writer = &fktestutils.CsvDataWriter{}
	} else if o.PostStream {
		writer = fktestutils.NewStreamingWriter(o.Host)
	} else if o.PostJson {
		writer = fktestutils.NewDataBinaryToPostWriter(o.Scheme, o.Host)
	} else if o.Split > 0 {
		writer = &fktestutils.SplittingWriter{
			Size: o.Split,
		}
	} else {
		writer = &fktestutils.NullWriter{}
	}

	transformer := &fktestutils.TransformerChain{
		Chain: []fktestutils.RecordTransformer{
			&fktestutils.MetadataSaver{},
			&fktestutils.ForceDeviceId{
				DeviceId: o.DeviceId,
			},
		},
	}

	log.Printf("Using %T", writer)

	df := &fktestutils.DataFile{
		Transformer: transformer,
		Verbose:     o.Verbose,
	}

	for _, path := range flag.Args() {
		log.Printf("Opening %s", path)

		df.ReadData(path, writer)
	}

	writer.Finished()

	log.Printf("Done (%d records)", df.NumberOfRecords)
}
