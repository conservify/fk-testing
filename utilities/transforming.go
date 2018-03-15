package utilities

import (
	"encoding/hex"
	"fmt"
	"log"

	pb "github.com/fieldkit/data-protocol"
)

type TransformChainFunc func(*DataFile, *pb.DataRecord) error

type RecordTransformer interface {
	TransformRecord(df *DataFile, record *pb.DataRecord, chain TransformChainFunc) error
}

type NoopTransform struct {
}

func (t *NoopTransform) TransformRecord(df *DataFile, record *pb.DataRecord, chain TransformChainFunc) error {
	chain(df, record)
	return nil
}

type TransformerChain struct {
	Chain []RecordTransformer
}

func (t *TransformerChain) Invoke(df *DataFile, record *pb.DataRecord, last TransformChainFunc, n int) error {
	following := last

	if n < len(t.Chain)-1 {
		following = func(df *DataFile, record *pb.DataRecord) error {
			return t.Invoke(df, record, last, n+1)
		}
	}

	return t.Chain[n].TransformRecord(df, record, following)
}

func (t *TransformerChain) TransformRecord(df *DataFile, record *pb.DataRecord, chain TransformChainFunc) error {
	if len(t.Chain) == 0 {
		return chain(df, record)
	}
	return t.Invoke(df, record, chain, 0)
}

type ForceDeviceId struct {
	DeviceId string
}

func (t *ForceDeviceId) TransformRecord(df *DataFile, record *pb.DataRecord, chain TransformChainFunc) error {
	if t.DeviceId == "" {
		return chain(df, record)
	}

	if record.Metadata != nil {
		deviceId, err := hex.DecodeString(t.DeviceId)
		if err != nil {
			return fmt.Errorf("Unable to decode DeviceId: %v", err)
		}

		log.Printf("Forcing DeviceId = %v (was %v)", deviceId, record.Metadata.DeviceId)

		record.Metadata.DeviceId = deviceId
	}

	return chain(df, record)
}
