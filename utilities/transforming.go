package utilities

import (
	"encoding/hex"
	"fmt"
	"log"

	pb "github.com/fieldkit/data-protocol"
)

type BeginChainFunc func(*DataFile) error
type ProcessChainFunc func(*DataFile, *pb.DataRecord) error
type EndChainFunc func(*DataFile) error

type RecordTransformer interface {
	Begin(df *DataFile, chain BeginChainFunc) error
	Process(df *DataFile, record *pb.DataRecord, begin BeginChainFunc, chain ProcessChainFunc, end EndChainFunc) error
	End(df *DataFile, chain EndChainFunc) error
}

type NoopTransform struct {
}

func (t *NoopTransform) Begin(df *DataFile, chain BeginChainFunc) error {
	return chain(df)
}

func (t *NoopTransform) Process(df *DataFile, record *pb.DataRecord, begin BeginChainFunc, chain ProcessChainFunc, end EndChainFunc) error {
	return chain(df, record)
}

func (t *NoopTransform) End(df *DataFile, chain EndChainFunc) error {
	return chain(df)
}

type ForceDeviceId struct {
	DeviceId string
}

func (t *ForceDeviceId) Begin(df *DataFile, chain BeginChainFunc) error {
	return chain(df)
}

func (t *ForceDeviceId) Process(df *DataFile, record *pb.DataRecord, begin BeginChainFunc, chain ProcessChainFunc, end EndChainFunc) error {
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

func (t *ForceDeviceId) End(df *DataFile, chain EndChainFunc) error {
	return chain(df)
}
