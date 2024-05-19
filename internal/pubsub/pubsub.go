package pubsub

import (
	"fmt"
	"log"
	"sync"

	"mist-to-tsdb/pkg/mistdatafmt"
)

type pubsubIntfBackend interface {
	PublishRaw(PubsubIntfTarget, string) error
	Close() error
}

type PubsubIntfConf struct {
	Debug			bool				`mapstructure:"debug",default:false`
	Driver			string				`mapstructure:"driver",default:"kafka"`
	DriverKafka		PubsubIntfConfDrvKafka		`mapstructure:"kafka"`
	Datasource		[]PubsubIntfTarget
	DataInChannel		chan mistdatafmt.WsMsgData
}

type GenericKV struct {
	Key		string	`mapstructure:"key"`
	Value		string	`mapstructure:"value"`
}

type PubsubIntfTarget struct {
	Channel		string
	Topic		string
	Header		[]GenericKV
}

type PubsubIntf struct {
	cfg		PubsubIntfConf
	backend		pubsubIntfBackend
	dataIn		chan mistdatafmt.WsMsgData
	wg		*sync.WaitGroup
	topicMap	map[string]PubsubIntfTarget
}

func New(cfg PubsubIntfConf) (*PubsubIntf, error) {
	var err error
	r := &PubsubIntf {
		cfg:		cfg,
		dataIn:		cfg.DataInChannel,
		topicMap:	make(map[string]PubsubIntfTarget),
	}

	for i := 0; i < len(cfg.Datasource); i++ {
		if cfg.Datasource[i].Channel == "" {
			return nil, fmt.Errorf("Missing channel in datasource")
		} else if cfg.Datasource[i].Topic == "" {
			return nil, fmt.Errorf("Missing topic for channel %s", cfg.Datasource[i].Channel)
		}

		r.topicMap[cfg.Datasource[i].Channel] = cfg.Datasource[i]
	}

	// Init Backend Driver
	switch cfg.Driver {
	case "kafka":
		r.backend, err = pubsubIntfKafkaNew(cfg)
		if err != nil {
			return nil, err
		}

	case "dummy":
		r.backend, err = pubsubIntfDummyNew(cfg)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("Unknown PubSub driver: %s", cfg.Driver)
	}

	return r, nil
}

func (i *PubsubIntf) Run(wg *sync.WaitGroup, killSig chan struct{}) error {
	var err error

	// Final pre-flight init
	i.wg = wg
	wg.Add(1)
	defer i.finish()

	// Main routine
	for {
		select {
		case <-killSig:
			return nil
		case msg := <-i.dataIn:
			if i.cfg.Debug {
				log.Printf("PubSub Start Process: %v", msg)
			}

			err = i.processData(msg.Channel, msg.Data)
			if err != nil {
				log.Printf("PubSub driver has thrown error: %v", err)
				log.Printf("Failed data: %v", msg)
			}
		}
	}

	return nil
}

func (i *PubsubIntf) processData(channel string, data string) error {
	tgt, e := i.topicMap[channel]
	if !e {
		return fmt.Errorf("Data received on channel %s, but topic not defined", channel)
	}
		
	err := i.backend.PublishRaw(tgt, data)
	return err
}

func (i *PubsubIntf) finish() {
	err := i.backend.Close()
	if err != nil {
		log.Printf("PubSub driver has thrown error on close: %v", err)
	}

	if i.wg != nil {
		i.wg.Done()
	}

	return
}
