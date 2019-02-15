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
	Streams []lokiStream `json:"streams"`
}

type lokiStream struct {
	Labels map[string]interface{} `json:"labels"`
	Entries []lokiEntry             `json:"entries"`
}

type lokiEntry struct {
	Ts string                     `json:"ts"`
	Line map[string]interface{}   `json:"line"`
}

func extractMetadata(finalServiceName string, message []byte) map[string]interface{} {

	logrus.WithField("service_name", finalServiceName).Debug("searching for extractor")

	metadata := map[string]interface{}{}

	if bridge, err := bridge.New(finalServiceName); err == nil {
		metadata = bridge.Extract(message)
	} else {
		logrus.WithField("service_name", finalServiceName).Debug("extractor not found")
		metadata = map[string]interface{}{
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

	if metadata["time"] != nil {
		if parsedTime, err := time.Parse(time.RFC3339, metadata["time"].(string)); err != nil {
			lp.logLine.Timestamp = parsedTime
		}

		delete(metadata, "time")
	}

	entryToSend := lokiEntry {
		Ts: lp.logLine.Timestamp.Format(time.RFC3339),
		Line: metadata,
	}

	streamToSend := lokiStream {
		Labels: structs.Map(lp.logLine),
		Entries: []lokiEntry{entryToSend},
	}

	lokiToSend := lokiMsg {
		Streams: []lokiStream{streamToSend},
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