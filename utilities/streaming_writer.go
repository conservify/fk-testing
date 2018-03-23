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
	host  string
	async bool
	buf   *proto.Buffer
}

func NewStreamingWriter(host string, async bool) *StreamingWriter {
	return &StreamingWriter{
		host:  host,
		async: async,
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

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(all))
	if err != nil {
		return fmt.Errorf("Error uploading %v", err)
	}

	req.Header.Add("Content-Type", "application/vnd.fk.data+binary")
	if w.async {
		req.Header.Add("Fk-Processing", "async")
	}

	cl := http.Client{}
	resp, err := cl.Do(req)
	if err != nil {
		return fmt.Errorf("Error uploading %v", err)
	}

	log.Printf("Done [%d] %s", resp.StatusCode, resp.Status)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	body := string(bodyBytes)

	if resp.StatusCode != 200 {
		return fmt.Errorf("Server error: (%v): %s", resp.StatusCode, body)
	}

	return chain(df)
}
