package utilities

import (
	"log"

	"github.com/golang/protobuf/proto"

	pb "github.com/fieldkit/data-protocol"
)

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
