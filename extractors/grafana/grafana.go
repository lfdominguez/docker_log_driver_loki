package grafana

import (
	"regexp"
	"strconv"

	"github.com/lfdominguez/docker_log_driver_loki/bridge"
)

type GrafanaAdapter struct{}

func init() {
	f := new(Factory)
	bridge.Register(f, "grafana")
}

func convert(value string) interface{} {
	if b, err := strconv.ParseBool(value); err != nil {
		return b
	}

	if f, err := strconv.ParseFloat(value, 64); err != nil {
		return f
	}

	if i, err := strconv.ParseInt(value, 10, 64); err != nil {
		return i
	}

	return value
}

var (
	logPattern = regexp.MustCompile(`([a-zA-Z])+=("([^"]+)"|[^ ]+)`)
)

type Factory struct{}

func (f *Factory) New(serviceName string) bridge.ExtractorAdapter {

	if serviceName == "grafana" {
		return &GrafanaAdapter{}
	}

	return nil
}

func (g *GrafanaAdapter) Extract(message []byte) map[string]interface{} {
	matches := logPattern.FindAllStringSubmatch(string(message), -1)

	ret := make(map[string]interface{})

	for _, match := range matches {
		key := match[1]
		value := match[3]

		if value == "" {
			value = match[2]
		}

		ret[key] = convert(value)
	}

	ret["level"] = ret["lvl"]
	ret["time"] = ret["t"]

	delete(ret, "t")
	delete(ret, "lvl")

	return ret
}
