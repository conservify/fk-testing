package utilities

import (
	_ "fmt"
	fkc "github.com/fieldkit/app-protocol/fkdevice"
	_ "gopkg.in/cheggaaa/pb.v1"
	_ "io"
	_ "log"
	_ "os"
	_ "time"
)

func DownloadDeviceFiles(dataDirectory string, deviceId string, dc *fkc.DeviceClient) (files []string, err error) {
	files = make([]string, 0)
	/*
		fileReply, err := dc.QueryFiles()
		if err != nil {
			return files, fmt.Errorf("Error: %v", err)
		}

		if fileReply == nil || fileReply.Files == nil || fileReply.Files.Files == nil {
			return files, fmt.Errorf("Error, no files in reply")
		}

		for _, file := range fileReply.Files.Files {
			if file.Size > 0 {
				now := time.Now()
				dir := fmt.Sprintf("%s/%s", dataDirectory, deviceId)
				stamp := now.Format("20060102_150405")
				yearMonth := now.Format("200601")
				day := now.Format("02")

				stampedDir := fmt.Sprintf("%s/%s/%s", dir, yearMonth, day)
				fileName := fmt.Sprintf("%s/%s_%s", stampedDir, file.Name, stamp)

				err = os.MkdirAll(stampedDir, 0777)
				if err != nil {
					return files, fmt.Errorf("Unable to create %s (%v)", stampedDir, err)
				}

				f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
				if err != nil {
					return files, fmt.Errorf("Unable to open %s (%v)", fileName, err)
				}

				defer f.Close()

				files = append(files, fileName)

				log.Printf("[%s] Downloading %v (%s)", deviceId, file, fileName)

				bar := pb.New(int(file.Size)).SetUnits(pb.U_BYTES)
				bar.Start()

				writer := io.MultiWriter(f, bar)

				err = dc.DownloadFileToWriter(file.Id, 0, 0, 0, writer)

				bar.Set(int(file.Size))
				bar.Finish()

				if err != nil {
					log.Printf("Deleting incomplete file %s", fileName)
					if removeErr := os.Remove(fileName); removeErr != nil {
						log.Printf("Error deleting incomplete file: %v", removeErr)
					}
					return files, fmt.Errorf("Error: %v", err)
				} else {
					_, err = dc.EraseFile(file.Id)
					if err != nil {
						return files, fmt.Errorf("Error erasing device file: %v", err)
					}
				}
			}
		}
	*/

	return files, nil
}
