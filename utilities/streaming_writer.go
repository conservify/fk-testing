package utilities

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/golang/protobuf/proto"

	pb "github.com/fieldkit/data-protocol"
)

type StreamingWriter struct {
	host string
	buf  *proto.Buffer
}

func NewStreamingWriter(host string) *StreamingWriter {
	return &StreamingWriter{
		host: host,
	}
}

func (w *StreamingWriter) Begin(df *DataFile, chain BeginChainFunc) error {
	w.buf = proto.NewBuffer(nil)

	return chain(df)
}

func (w *StreamingWriter) Process(df *DataFile, record *pb.DataRecord, begin BeginChainFunc, chain ProcessChainFunc, end EndChainFunc) error {
	raw, err := df.Marshal(record)
	if err != nil {
		return fmt.Errorf("Error writing to streaming writer: %v", err)
	}

	w.buf.EncodeRawBytes(raw)

	return chain(df, record)
}

func (w *StreamingWriter) End(df *DataFile, chain EndChainFunc) error {
	all := w.buf.Bytes()

	url := fmt.Sprintf("http://%s/messages/ingestion/stream", w.host)

	log.Printf("Connecting to %s and uploading %d bytes", url, len(all))

	r, err := http.Post(url, "application/vnd.fk.data+binary", bytes.NewBuffer(all))
	if err != nil {
		return fmt.Errorf("Error uploading %v", err)
	}

	log.Printf("Done [%d] %s", r.StatusCode, r.Status)

	bodyBytes, err := ioutil.ReadAll(r.Body)
	body := string(bodyBytes)

	if r.StatusCode != 200 {
		return fmt.Errorf("Server error: (%v): %s", r.StatusCode, body)
	}

	return chain(df)
}
