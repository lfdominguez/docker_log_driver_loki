package bridge

type ExtractorAdapter interface {
	Extract(message []byte) map[string]string
}

type AdapterFactory interface {
	New(serviceName string) ExtractorAdapter
}