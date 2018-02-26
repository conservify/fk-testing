package utilities

import (
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
