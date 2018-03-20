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

	SplitRecords int
	SplitBytes   int

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

	flag.IntVar(&o.SplitBytes, "split-bytes", 0, "split stream into smaller batches")
	flag.IntVar(&o.SplitRecords, "split-records", 0, "split stream into smaller batches")

	flag.StringVar(&o.Project, "project", "www", "project")
	flag.StringVar(&o.DeviceName, "device-name", "weather-proxy", "device name")
	flag.StringVar(&o.Scheme, "scheme", "http", "scheme to use")
	flag.StringVar(&o.Host, "host", "127.0.0.1:8080", "hostname to use")
	flag.StringVar(&o.Username, "username", "demo-user", "username to use")
	flag.StringVar(&o.Password, "password", "asdfasdfasdf", "password to use")

	flag.StringVar(&o.DeviceId, "force-device-id", "", "force uploaded data to use this device id")

	flag.Parse()

	chain := make([]fktestutils.RecordTransformer, 0)
	chain = append(chain, &fktestutils.MetadataSaver{})
	chain = append(chain, &fktestutils.ForceDeviceId{
		DeviceId: o.DeviceId,
	})

	if o.SplitBytes > 0 || o.SplitRecords > 0 {
		chain = append(chain, &fktestutils.SplittingWriter{
			AfterBytes:   o.SplitBytes,
			AfterRecords: o.SplitRecords,
		})
	}

	if o.Csv {
		chain = append(chain, &fktestutils.CsvDataWriter{})
	} else if o.PostStream {
		chain = append(chain, fktestutils.NewStreamingWriter(o.Host))
	} else if o.PostJson {
		chain = append(chain, fktestutils.NewDataBinaryToPostWriter(o.Scheme, o.Host))
	} else {
		chain = append(chain, &fktestutils.NullWriter{})
	}

	transformer := &fktestutils.TransformerChain{
		Chain: chain,
	}

	log.Printf("Using %#v", transformer)

	df := &fktestutils.DataFile{
		Transformer: transformer,
		Verbose:     o.Verbose,
	}

	for _, path := range flag.Args() {
		log.Printf("Opening %s", path)

		df.ReadData(path)
	}

	log.Printf("Done (%d records)", df.NumberOfRecords)
}
