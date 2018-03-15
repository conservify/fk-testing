package utilities

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/golang/protobuf/proto"

	pb "github.com/fieldkit/data-protocol"
)

const (
	MetadataFilename = "Metadata.fkpb"
)

type StreamingWriter struct {
	host     string
	response *http.Response
	buf      *proto.Buffer
}

func NewStreamingWriter(host string) *StreamingWriter {
	return &StreamingWriter{
		host: host,
	}
}

func (w *StreamingWriter) WriteRecord(raw []byte) {
	if w.buf == nil {
		w.buf = proto.NewBuffer(nil)
	}

	w.buf.EncodeRawBytes(raw)
}

func (w *StreamingWriter) Write(df *DataFile, record *pb.DataRecord) error {
	raw, err := df.Marshal(record)
	if err != nil {
		return fmt.Errorf("Error writing to streaming writer: %v", err)
	}

	w.WriteRecord(raw)
	return nil
}

func (w *StreamingWriter) Finished() error {
	all := w.buf.Bytes()

	url := fmt.Sprintf("http://%s/messages/ingestion/stream", w.host)

	log.Printf("Connecting to %s and uploading %d bytes", url, len(all))

	c, err := http.Post(url, "application/vnd.fk.data+binary", bytes.NewBuffer(all))
	if err != nil {
		return err
	}

	w.response = c

	return nil
}
