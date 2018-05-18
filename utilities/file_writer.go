package utilities

import (
	"fmt"
	"log"
	"os"

	"github.com/golang/protobuf/proto"

	pb "github.com/fieldkit/data-protocol"
)

type FileWriter struct {
	Name     string
	Numbered bool
	Number   int
	File     *os.File
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

func (fw *FileWriter) GenerateFileName() (string, error) {
	i := 0
	fn := fw.Name

	for {
		if fw.Numbered {
			fn = fmt.Sprintf("%s_%d", fw.Name, i)
		}

		if _, err := os.Stat(fn); err == nil {
			if !fw.Numbered {
				return "", fmt.Errorf("File already exists: %s", fn)
			}
			i += 1
		} else {
			return fn, nil
		}
	}
}

func (fw *FileWriter) Begin(df *DataFile, chain BeginChainFunc) error {
	fn, err := fw.GenerateFileName()
	if err != nil {
		return err
	}

	log.Printf("Opening %s...", fn)

	f, err := os.Create(fn)
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
