package driver

import (
	"bytes"
	"encoding/json"
	"time"
	"net/http"

	"github.com/lfdominguez/docker_log_driver_loki/extractors"
	"github.com/fatih/structs"
)

type lokiMsg struct {
	streams []lokiStream
}

type lokiStream struct {
	labels []map[string]interface{}
	entries []lokiEntry
}

type lokiEntry struct {
	ts string
	line map[string]string
}

func extractMetadata(finalServiceName string, message []byte) map[string]string {

	bridge, err := extractors.New(finalServiceName)

	if err != nil {
		return map[string]string{}
	}

	return bridge.Extract(message)
}

func logMessageToLoki(lp *logPair, message []byte) error {
	stackStrSize := len(lp.logLine.StackName)

	// Convert to Runes to avoid problems with multibytes (Unicode, UTF-8, UTF-16, etc) characters.
	runes := []rune(lp.logLine.ServiceName)

	finalServiceName := runes[stackStrSize:]

	metadata := extractMetadata(string(finalServiceName), message)

	lp.logLine.Timestamp = time.Now()

	if len(metadata) != 0 {

		if metadata["time"] != "" {
			if parsedTime, err := time.Parse(time.RFC3339, metadata["time"]); err != nil {
				lp.logLine.Timestamp = parsedTime
			}

			delete(metadata, "time")
		}
	}

	if metadata["message"] == "" {
		metadata["message"] = string(message[:])
	}

	entryToSend := lokiEntry {
		ts: lp.logLine.Timestamp.Format(time.RFC3339),
		line: metadata,
	}

	streamToSend := lokiStream {
		labels: []map[string]interface{}{structs.Map(lp.logLine)},
		entries: []lokiEntry{entryToSend},
	}

	lokiToSend := lokiMsg {
		streams: []lokiStream{streamToSend},
	}

	bytesToSend, err := json.Marshal(lokiToSend)
	if err != nil {
		return err
	}
	
	url := "http://192.168.0.159/api/prom/push"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bytesToSend))
	
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
	
	resp, err := client.Do(req)	
    if err != nil {
        return err
	}
	
    defer resp.Body.Close()

	return nil
}