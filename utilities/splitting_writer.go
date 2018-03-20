package utilities

import (
	_ "fmt"
	"log"
	"os"

	"github.com/golang/protobuf/proto"

	pb "github.com/fieldkit/data-protocol"
)

type FileWriter struct {
	Name string
	File *os.File
}

func (fw *FileWriter) Append(df *DataFile, record *pb.DataRecord) error {
	bytes, err := df.Marshal(record)
	if err != nil {
		return err
	}

	buf := proto.NewBuffer(nil)
	buf.EncodeRawBytes(bytes)

	_, err = fw.File.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (fw *FileWriter) Begin(df *DataFile, chain BeginChainFunc) error {
	if _, err := os.Stat(fw.Name); err == nil {
		log.Fatalf("File already exists: %s", fw.Name)
	}

	log.Printf("Opening %s...", fw.Name)

	f, err := os.Create(fw.Name)
	if err != nil {
		return err
	}

	fw.File = f

	return chain(df)
}

func (fw *FileWriter) Process(df *DataFile, record *pb.DataRecord, begin BeginChainFunc, chain ProcessChainFunc, end EndChainFunc) error {
	err := fw.Append(df, record)
	if err != nil {
		return err
	}
	return chain(df, record)
}

func (fw *FileWriter) End(df *DataFile, chain EndChainFunc) error {
	if fw.File != nil {
		fw.File.Close()
		fw.File = nil
	}

	return chain(df)
}

type SplittingWriter struct {
	AfterBytes       int
	AfterRecords     int
	BytesProcessed   int
	RecordsProcessed int
	Sequence         int
	Metadata         *pb.DataRecord
}

func (sw *SplittingWriter) Begin(df *DataFile, chain BeginChainFunc) error {
	sw.BytesProcessed = 0
	sw.RecordsProcessed = 0
	return chain(df)
}

func (sw *SplittingWriter) shouldSplit(record *pb.DataRecord) bool {
	if record.LoggedReading != nil && record.LoggedReading.Reading != nil {
		reading := record.LoggedReading.Reading
		if reading.Sensor > 0 {
			return false
		}
	}

	if sw.AfterBytes > 0 {
		return sw.BytesProcessed >= sw.AfterBytes
	}

	if sw.AfterRecords > 0 {
		return sw.RecordsProcessed >= sw.AfterRecords
	}

	return false
}

func (sw *SplittingWriter) Process(df *DataFile, record *pb.DataRecord, begin BeginChainFunc, chain ProcessChainFunc, end EndChainFunc) error {
	if IsAcceptableMetadataRecord(record) {
		sw.Metadata = record
	}

	if sw.shouldSplit(record) {
		err := end(df)
		if err != nil {
			return err
		}

		err = begin(df)
		if err != nil {
			return err
		}

		sw.BytesProcessed = 0
		sw.RecordsProcessed = 0

		if sw.Metadata != nil {
			log.Printf("(SplittingWriter) Appending metadata (%d).", len(sw.Metadata.Metadata.Sensors))
			err := chain(df, sw.Metadata)
			if err != nil {
				return err
			}
		} else {
			log.Printf("(SplittingWriter) No metadata.")
		}

	}

	bytes, err := df.Marshal(record)
	if err != nil {
		return err
	}

	buf := proto.NewBuffer(nil)
	buf.EncodeRawBytes(bytes)

	sw.BytesProcessed += len(buf.Bytes())
	sw.RecordsProcessed += 1

	return chain(df, record)
}

func (sw *SplittingWriter) End(df *DataFile, chain EndChainFunc) error {
	return chain(df)
}
