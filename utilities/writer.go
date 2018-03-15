package utilities

import (
	"io"
	"io/ioutil"
	"log"

	"github.com/golang/protobuf/proto"

	pb "github.com/fieldkit/data-protocol"
)

type DataFile struct {
	Path            string
	Verbose         bool
	Transformer     RecordTransformer
	NumberOfRecords int
	LastMetadata    *pb.DataRecord
}

func (df *DataFile) Unmarshal(raw []byte) (record *pb.DataRecord, err error) {
	buffer := proto.NewBuffer(raw[:])
	record = new(pb.DataRecord)
	err = buffer.Unmarshal(record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (df *DataFile) Marshal(record *pb.DataRecord) (raw []byte, err error) {
	raw, err = proto.Marshal(record)
	if err != nil {
		return nil, err
	}

	return raw, nil
}

func (df *DataFile) ReadData(path string, dw DataWriter) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}

	df.Path = path

	buf := proto.NewBuffer(data[:])

	for {
		messageBytes, err := buf.DecodeRawBytes(true)
		if err != nil {
			if err == io.ErrUnexpectedEOF {
				err = nil
				break
			}
			log.Fatalf("%v", err)
		}

		record, err := df.Unmarshal(messageBytes)
		if err != nil {
			log.Printf("Unable to unmarshal from file: %v (%d bytes)", err, len(messageBytes))
		} else {
			if df.Verbose {
				log.Printf("%+v", record)
			}

			last := func(df *DataFile, record *pb.DataRecord) error {
				if record.Metadata != nil {
					df.LastMetadata = record
				}

				return dw.Write(df, record)
			}

			err = df.Transformer.TransformRecord(df, record, last)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}

			df.NumberOfRecords += 1
		}
	}
}

type DataWriter interface {
	Write(df *DataFile, record *pb.DataRecord) error
	Finished() error
}

type NullWriter struct {
}

func (w *NullWriter) Write(df *DataFile, record *pb.DataRecord) error {
	return nil
}

func (w *NullWriter) Finished() error {
	return nil
}
