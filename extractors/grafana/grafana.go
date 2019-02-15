package grafana

import (
	"regexp"

	"github.com/lfdominguez/docker_log_driver_loki/bridge"
)

type GrafanaAdapter struct {}

func init() {
	f := new(Factory)
	bridge.Register(f, "grafana")
}

var (
	logPattern = regexp.MustCompile(`([a-zA-Z])+=("([^\"]+)"|[^\ ]+)`)
)

type Factory struct{}

func (f *Factory) New(serviceName string) bridge.ExtractorAdapter {

	if serviceName == "grafana" {
		return &GrafanaAdapter{}
	}

	return nil
}

func (g *GrafanaAdapter) Extract(message []byte) map[string]string {
	matches := logPattern.FindAllStringSubmatch(string(message), -1)

	ret := make(map[string]string)

	for _, match := range matches {
		key := match[1]
		value := match[3]
		if value == "" {
			value = match[2]
		}
		ret[key] = value
	}

	ret["level"] = ret["lvl"]
	ret["time"] = ret["t"]

	delete(ret, "t")
	delete(ret, "lvl")

	return ret
}