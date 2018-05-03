package utilities

import (
	"fmt"
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

func (df *DataFile) ReadData(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}

	df.Path = path

	position := 0

	lastBegin := func(*DataFile) error {
		return nil
	}
	lastProcess := func(df *DataFile, record *pb.DataRecord) error {
		if record.Metadata != nil {
			df.LastMetadata = record
		}
		return nil
	}
	lastEnd := func(*DataFile) error {
		return nil
	}

	err = df.Transformer.Begin(df, lastBegin)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

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

		temp := proto.EncodeVarint(uint64(len(messageBytes)))
		position += len(messageBytes) + len(temp)
		df.NumberOfRecords += 1

		record, err := df.Unmarshal(messageBytes)
		if err != nil {
			log.Printf("Unable to unmarshal from file: %v (%d bytes)", err, len(messageBytes))
		} else {
			if df.Verbose {
				log.Printf("(%d) %+v", position, record)
			}

			err = df.Transformer.Process(df, record, lastBegin, lastProcess, lastEnd)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
		}
	}

	err = df.Transformer.End(df, lastEnd)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	return nil
}

type NullWriter struct {
	Processed int
}

func (w *NullWriter) Begin(df *DataFile, chain BeginChainFunc) error {
	w.Processed = 0
	return chain(df)
}

func (w *NullWriter) Process(df *DataFile, record *pb.DataRecord, begin BeginChainFunc, chain ProcessChainFunc, end EndChainFunc) error {
	w.Processed += 1
	return chain(df, record)
}

func (w *NullWriter) End(df *DataFile, chain EndChainFunc) error {
	log.Printf("(NullWriter) End, processed = %d", w.Processed)
	return chain(df)
}

type LogWriter struct {
	Processed int
}

func (w *LogWriter) Begin(df *DataFile, chain BeginChainFunc) error {
	return chain(df)
}

func (w *LogWriter) Process(df *DataFile, record *pb.DataRecord, begin BeginChainFunc, chain ProcessChainFunc, end EndChainFunc) error {
	if record.Log != nil {
		fmt.Printf("%-10d %-30s %s\n", record.Log.Uptime, record.Log.Facility, record.Log.Message)
	}
	return chain(df, record)
}

func (w *LogWriter) End(df *DataFile, chain EndChainFunc) error {
	return chain(df)
}
