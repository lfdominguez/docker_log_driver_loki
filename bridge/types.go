package bridge

type ExtractorAdapter interface {
	Extract(message []byte) map[string]interface{}
}

type AdapterFactory interface {
	New(serviceName string) ExtractorAdapter
}