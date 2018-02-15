package utilities

import (
	"fmt"
	fkc "github.com/fieldkit/app-protocol/fkdevice"
	"gopkg.in/cheggaaa/pb.v1"
	"io"
	"log"
	"os"
	"time"
)

func DownloadDeviceFiles(dataDirectory string, deviceId string, dc *fkc.DeviceClient) (files []string, err error) {
	files = make([]string, 0)
	fileReply, err := dc.QueryFiles()
	if err != nil {
		return files, fmt.Errorf("Error: %v", err)
	}

	if fileReply == nil || fileReply.Files == nil || fileReply.Files.Files == nil {
		return files, fmt.Errorf("Error, no files in reply")
	}

	for _, file := range fileReply.Files.Files {
		if file.Size > 0 {
			dir := fmt.Sprintf("%s/%s", dataDirectory, deviceId)
			stamp := time.Now().Format("20060102_150405")
			fileName := fmt.Sprintf("%s/%s_%s_%d", dir, file.Name, stamp, file.Version)

			err = os.MkdirAll(dir, 0777)
			if err != nil {
				return files, fmt.Errorf("Unable to create %s (%v)", dir, err)
			}

			f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
			if err != nil {
				return files, fmt.Errorf("Unable to open %s (%v)", fileName, err)
			}

			files = append(files, fileName)

			defer f.Close()

			log.Printf("[%s] Downloading %v (%s)", deviceId, file, fileName)

			bar := pb.New(int(file.Size)).SetUnits(pb.U_BYTES)
			bar.Start()

			writer := io.MultiWriter(f, bar)

			token := []byte{}

			err = dc.DownloadFileToWriter(file.Id, 65536*1, token, writer)

			bar.Set(int(file.Size))
			bar.Finish()

			if err != nil {
				return files, fmt.Errorf("Error: %v", err)
			} else {
				dc.EraseFile(file.Id)
				if err != nil {
					return files, fmt.Errorf("Error: %v", err)
				}
			}
		}
	}

	return files, nil
}
