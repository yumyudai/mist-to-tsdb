package pubsub

import (
	"log"
)

type pubsubIntfDummy struct{}

func pubsubIntfDummyNew(_ PubsubIntfConf) (*pubsubIntfDummy, error) {
	// start building interface
	r := &pubsubIntfDummy{}
	return r, nil
}

func (s *pubsubIntfDummy) Close() error {
	return nil
}

func (s *pubsubIntfDummy) PublishRaw(target PubsubIntfTarget, data string) error {
	log.Printf("Publish Target %v Data %s", target, data)	
	return nil
}

