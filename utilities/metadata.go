package utilities

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	pb "github.com/fieldkit/data-protocol"
)

type MetadataSaver struct {
	WroteMetadata bool
}

func (w *MetadataSaver) TransformRecord(df *DataFile, record *pb.DataRecord, chain TransformChainFunc) error {
	if !w.WroteMetadata {
		if _, err := os.Stat(MetadataFilename); err == nil {
			bytes, err := ioutil.ReadFile(MetadataFilename)
			if err != nil {
				return err
			}

			log.Printf("Inserting metadata (%d bytes).", len(bytes))

			metadata, err := df.Unmarshal(bytes)
			if err != nil {
				return fmt.Errorf("Unable to unmarshal saved metadata: %v", err)
			}

			err = chain(df, metadata)
			if err != nil {
				return err
			}
		} else {
			log.Printf("No saved metadata.")
		}

		w.WroteMetadata = true
	}

	if record.Metadata != nil && record.Metadata.Sensors != nil && len(record.Metadata.Sensors) > 0 {
		bytes, err := df.Marshal(record)
		if err != nil {
			return fmt.Errorf("Unable to marshal metadata: %v", err)
		}

		err = ioutil.WriteFile(MetadataFilename, bytes, 0644)
		if err != nil {
			return fmt.Errorf("Unable to save metadata: %v", err)
		}

		log.Printf("Saved metadata (%d bytes).", len(bytes))
	}

	return chain(df, record)
}
