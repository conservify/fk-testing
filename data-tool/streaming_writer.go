package main

import (
	"bytes"
	"fmt"
	pb "github.com/fieldkit/data-protocol"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const (
	MetadataFilename = "Metadata.fkpb"
)

type StreamingWriter struct {
	options      *options
	response     *http.Response
	buf          *proto.Buffer
	haveMetadata bool
}

func (w *StreamingWriter) WriteRecord(raw []byte) {
	if w.buf == nil {
		w.buf = proto.NewBuffer(nil)
	}

	w.buf.EncodeRawBytes(raw)
}

func (w *StreamingWriter) Write(df *DataFile, record *pb.DataRecord, raw []byte) error {
	if record.Metadata != nil {
		log.Printf("Saving metadata")
		ioutil.WriteFile(MetadataFilename, raw, 0644)
		w.haveMetadata = true
	} else {
		if !w.haveMetadata {
			if _, err := os.Stat(MetadataFilename); err == nil {
				b, err := ioutil.ReadFile(MetadataFilename)
				if err != nil {
					log.Fatal(err)
				}
				_ = b
				log.Printf("Writing saved metadtaa")
				w.haveMetadata = true
				w.WriteRecord(b)
			}
		}

		w.WriteRecord(raw)
	}

	return nil
}

func (w *StreamingWriter) Finished() error {
	all := w.buf.Bytes()

	url := fmt.Sprintf("http://%s/messages/ingestion/stream", w.options.Host)
	log.Printf("Connecting to %s and uploading %d bytes", url, len(all))

	c, err := http.Post(url, "application/vnd.fk.data+binary", bytes.NewBuffer(all))
	if err != nil {
		log.Fatalf("%v", err)
	}

	w.response = c
	return nil
}
