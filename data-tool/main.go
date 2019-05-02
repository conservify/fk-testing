package main

import (
	"flag"
	"log"

	fktestutils "github.com/fieldkit/testing/utilities"
)

type options struct {
	Csv             bool
	PostJson        bool
	Verbose         bool
	PostStreamSync  bool
	PostStreamAsync bool

	Log bool

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

	File string
}

func main() {
	o := options{}

	flag.BoolVar(&o.Csv, "csv", false, "write csv")
	flag.BoolVar(&o.PostJson, "post-json", false, "interpret and post json")
	flag.BoolVar(&o.PostStreamSync, "post-stream-sync", false, "post binary stream and process synchronously")
	flag.BoolVar(&o.PostStreamAsync, "post-stream-async", false, "post binary stream and process asynchronously")
	flag.BoolVar(&o.Verbose, "verbose", false, "increased verbosity")

	flag.BoolVar(&o.Log, "log", false, "display log")

	flag.IntVar(&o.SplitBytes, "split-bytes", 0, "split stream into smaller batches")
	flag.IntVar(&o.SplitRecords, "split-records", 0, "split stream into smaller batches")

	flag.StringVar(&o.Project, "project", "www", "project")
	flag.StringVar(&o.DeviceName, "device-name", "weather-proxy", "device name")
	flag.StringVar(&o.Scheme, "scheme", "http", "scheme to use")
	flag.StringVar(&o.Host, "host", "127.0.0.1:8080", "hostname to use")
	flag.StringVar(&o.Username, "username", "demo-user", "username to use")
	flag.StringVar(&o.Password, "password", "asdfasdfasdf", "password to use")

	flag.StringVar(&o.File, "file", "", "write to a file")

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
		chain = append(chain, &fktestutils.CsvDataWriter{
			FormattedTimes: true,
		})
	} else if o.PostStreamSync {
		chain = append(chain, fktestutils.NewStreamingWriter(o.Host, false))
	} else if o.PostStreamAsync {
		chain = append(chain, fktestutils.NewStreamingWriter(o.Host, true))
	} else if o.PostJson {
		chain = append(chain, fktestutils.NewDataBinaryToPostWriter(o.Scheme, o.Host))
	} else if o.Log {
		chain = append(chain, &fktestutils.LogWriter{})
	} else if o.File != "" {
		chain = append(chain, &fktestutils.FileWriter{
			Name:     o.File,
			Numbered: true,
		})
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

		if err := df.ReadData(path); err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("Done (%d records) (%d bytes)", df.NumberOfRecords, df.BytesRead)
}
