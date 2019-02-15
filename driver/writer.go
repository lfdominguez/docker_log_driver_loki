package driver

import (
	"bytes"
	"encoding/json"
	"time"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/lfdominguez/docker_log_driver_loki/bridge"
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

	logrus.WithField("service_name", finalServiceName).Debug("searching for extractor")

	metadata := map[string]string{}

	if bridge, err := bridge.New(finalServiceName); err == nil {
		metadata = bridge.Extract(message)
	} else {
		logrus.WithField("service_name", finalServiceName).Debug("extractor not found")
		metadata = map[string]string{
			"msg": string(message),
		}
	}

	return metadata
}

func logMessageToLoki(lp *logPair, message []byte) error {
	stackStrSize := len(lp.logLine.StackName)

	// Convert to Runes to avoid problems with multibytes (Unicode, UTF-8, UTF-16, etc) characters.
	runes := []rune(lp.logLine.ServiceName)

	finalServiceName := runes[stackStrSize:]

	metadata := extractMetadata(string(finalServiceName), message)

	lp.logLine.Timestamp = time.Now()

	if metadata["time"] != "" {
		if parsedTime, err := time.Parse(time.RFC3339, metadata["time"]); err != nil {
			lp.logLine.Timestamp = parsedTime
		}

		delete(metadata, "time")
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

	var urlBuilder strings.Builder

	urlBuilder.WriteString("http://")
	urlBuilder.WriteString(lp.info.Config[lokihost])
	urlBuilder.WriteString(":")
	urlBuilder.WriteString(lp.info.Config[lokiport])
	urlBuilder.WriteString("/api/prom/push")
	
	url := urlBuilder.String()

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bytesToSend))
	
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
	
	resp, err := client.Do(req)	
    if err != nil {
        return err
	}
	
    defer resp.Body.Close()

	return nil
}