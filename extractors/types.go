package extractors

type ExtractorAdapter interface {
	extract(message []byte) map[string]string
}

type AdapterFactory interface {
	New(serviceName string) ExtractorAdapter
}