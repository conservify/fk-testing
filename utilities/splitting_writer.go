package utilities

import (
	"fmt"
	"log"
	"os"

	"github.com/golang/protobuf/proto"

	pb "github.com/fieldkit/data-protocol"
)

type SplittingWriter struct {
	Size         int
	File         *os.File
	FileNumber   int
	BytesWritten int
}

func (sw *SplittingWriter) Open(df *DataFile) error {
	if sw.File != nil {
		sw.File.Close()
		sw.FileNumber += 1
	}

	name := fmt.Sprintf("batch-%03d.fkpb", sw.FileNumber)
	if _, err := os.Stat(name); err == nil {
		log.Fatalf("File already exists: %s", name)
	}

	log.Printf("Opening %s...", name)

	f, err := os.Create(name)
	if err != nil {
		return err
	}

	sw.File = f
	sw.BytesWritten = 0

	if df.LastMetadata != nil {
		log.Printf("Appending last metadata.")

		err = sw.Append(df, df.LastMetadata)
		if err != nil {
			return err
		}
	}

	return nil
}

func (sw *SplittingWriter) Append(df *DataFile, record *pb.DataRecord) error {
	bytes, err := df.Marshal(record)
	if err != nil {
		return err
	}

	buf := proto.NewBuffer(nil)
	buf.EncodeRawBytes(bytes)

	_, err = sw.File.Write(buf.Bytes())
	if err != nil {
		return err
	}

	sw.BytesWritten += len(bytes)

	return nil
}

func (sw *SplittingWriter) Write(df *DataFile, record *pb.DataRecord) error {
	if sw.File == nil || sw.BytesWritten >= sw.Size {
		err := sw.Open(df)
		if err != nil {
			return err
		}
	}

	return sw.Append(df, record)
}

func (sw *SplittingWriter) Finished() error {
	if sw.File != nil {
		sw.File.Close()
		sw.File = nil
	}

	return nil
}
