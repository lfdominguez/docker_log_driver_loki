package extractors

import (
	"regexp"
)

type Factory struct{}

type GrafanaAdapter struct {}

func init() {
	f := new(Factory)
	Register(f, "grafana")
}

var (
	logPattern = regexp.MustCompile(`([a-zA-Z])+=("([^\"]+)"|[^\ ]+)`)
)

func (f *Factory) New(serviceName string) ExtractorAdapter {

	if serviceName == "grafana" {
		return &GrafanaAdapter{}
	}

	return nil
}

func (g *GrafanaAdapter) extract(message []byte) map[string]string {
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
	ret["message"] = ret["msg"]

	delete(ret, "t")
	delete(ret, "lvl")
	delete(ret, "msg")

	return ret
}