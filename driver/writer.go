package driver

import (
	"bytes"
	"encoding/json"
	"github.com/fatih/structs"
	"net/http"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/lfdominguez/docker_log_driver_loki/bridge"
)

type lokiMsg struct {
	Streams []lokiStream `json:"streams"`
}

type lokiStream struct {
	Labels  string      `json:"labels"`
	Entries []lokiEntry `json:"entries"`
}

type lokiEntry struct {
	Ts   string `json:"ts"`
	Line string `json:"line"`
}

func extractMetadata(finalServiceName string, message []byte) map[string]interface{} {

	logrus.WithField("service_name", finalServiceName).Debug("searching for extractor")

	metadata := map[string]interface{}{}

	if pluginBridge, err := bridge.New(finalServiceName); err == nil {
		metadata = pluginBridge.Extract(message)
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

	var labels []string

	for key, val := range metadata {
		var lineStr strings.Builder

		lineStr.WriteString(key)
		lineStr.WriteString("=\"")

		val = strings.Replace(val.(string), "'", "\\'", -1)
		val = strings.Replace(val.(string), "\"", "\\\"", -1)
		lineStr.WriteString(val.(string))

		lineStr.WriteString("\"")

		labels = append(labels, lineStr.String())
	}

	entryToSend := lokiEntry{
		Ts:   lp.logLine.Timestamp.Format(time.RFC3339),
		Line: "{" + strings.Join(labels, ",") + "}",
	}

	labels = labels[:0]

	for key, val := range structs.Map(lp.logLine) {
		if key == "Extra" || key == "Timestamp" {
			continue
		}

		var lineStr strings.Builder

		lineStr.WriteString(key)
		lineStr.WriteString("=\"")

		val = strings.Replace(val.(string), "'", "\\'", -1)
		val = strings.Replace(val.(string), "\"", "\\\"", -1)
		lineStr.WriteString(val.(string))

		lineStr.WriteString("\"")

		labels = append(labels, lineStr.String())
	}

	streamToSend := lokiStream{
		Labels:  "{" + strings.Join(labels, ",") + "}",
		Entries: []lokiEntry{entryToSend},
	}

	lokiToSend := lokiMsg{
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
