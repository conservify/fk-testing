package utilities

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"time"

	"github.com/golang/protobuf/proto"

	pb "github.com/fieldkit/data-protocol"
)

type DataFile struct {
	Path            string
	Verbose         bool
	Transformer     RecordTransformer
	LastMetadata    *pb.DataRecord
	NumberOfRecords int
	BytesRead       int
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
		return fmt.Errorf("Error opening file: %v", err)
	}

	log.Printf("Opened %s (%d bytes)", path, len(data))

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
		return err
	}

	buf := proto.NewBuffer(data[:])

	var previous *pb.DataRecord

	for {
		messageBytes, err := buf.DecodeRawBytes(true)
		if err != nil {
			if err == io.ErrUnexpectedEOF {
				err = nil

				if position < len(data) {
					tempBuffer := proto.NewBuffer(data[position:])
					expected, err := tempBuffer.DecodeVarint()
					if err != nil {
						return fmt.Errorf("Error getting final message length: %v", err)
					}
					log.Printf("Unexpected EOF: position = %d message = %d actual = %d", position, expected, len(data)-position)
				}

				break
			}
			return fmt.Errorf("Error: %v", err)
		}

		temp := proto.EncodeVarint(uint64(len(messageBytes)))
		lengthWithPrefix := len(messageBytes) + len(temp)

		df.NumberOfRecords += 1
		df.BytesRead += lengthWithPrefix

		record, err := df.Unmarshal(messageBytes)
		if err != nil {
			log.Printf("(%d) Unable to unmarshal from file: %v (%d bytes)", position, err, lengthWithPrefix)
		} else {
			if df.Verbose {
				if false {
					log.Printf("(%d) %+v", position, record)
				}

				if previous != nil {
					if previous.Readings.Reading != record.Readings.Reading-1 {
						panic("weird record number jump")
					}
					elapsed := int64(record.Readings.Time) - int64(previous.Readings.Time)
					if true || 30-elapsed > 10 || 30-elapsed < -10 {
						// locationTime := time.Unix(int64(previous.Readings.Location.Time), 0)
						previousTime := time.Unix(int64(previous.Readings.Time), 0)
						currentTime := time.Unix(int64(record.Readings.Time), 0)
						// log.Printf("(%d) %+v", position, record)
						log.Printf("(%d) #%d %s - #%d %s = %d",
							record.Readings.Reading-previous.Readings.Reading,
							previous.Readings.Reading,
							previousTime,
							record.Readings.Reading,
							currentTime,
							elapsed)
					}
				}
			}

			err = df.Transformer.Process(df, record, lastBegin, lastProcess, lastEnd)
			if err != nil {
				log.Printf("Error")
				return err
			}

			previous = record
		}

		position += lengthWithPrefix
	}

	err = df.Transformer.End(df, lastEnd)
	if err != nil {
		return err
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
	log.Printf("(NullWriter) End, records = %d", w.Processed)
	return chain(df)
}

var (
	LogLevelNames = map[uint32]string{
		0: "TRACE",
		1: "DEBUG",
		2: "INFO",
		3: "WARN",
		4: "ERROR",
	}
)

type LogWriter struct {
	Processed int
}

func (w *LogWriter) Begin(df *DataFile, chain BeginChainFunc) error {
	return chain(df)
}

const ()

func (w *LogWriter) Process(df *DataFile, record *pb.DataRecord, begin BeginChainFunc, chain ProcessChainFunc, end EndChainFunc) error {
	if record.Log != nil {
		fmt.Printf("%-10d %-10d %8s %-30s %s\n", record.Log.Uptime, record.Log.Time, LogLevelNames[record.Log.Level], record.Log.Facility, record.Log.Message)
	}
	return chain(df, record)
}

func (w *LogWriter) End(df *DataFile, chain EndChainFunc) error {
	return chain(df)
}
