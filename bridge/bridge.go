package bridge

import (
	"sync"
	"errors"
)

type Bridge struct {
	sync.Mutex
	registry       ExtractorAdapter
}

func New(serviceName string) (*Bridge, error) {
	factory, found := AdapterFactories.Lookup(serviceName)
	if !found {
		return nil, errors.New("Not found extractor for: " + serviceName)
	}

	return &Bridge{
		registry:       factory.New(serviceName),
	}, nil
}

func (b *Bridge) Extract(message []byte) map[string]interface{} {
	return b.registry.Extract(message)
}