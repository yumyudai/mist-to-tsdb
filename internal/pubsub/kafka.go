package pubsub

import (
	"fmt"
	"log"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type PubsubIntfConfDrvKafka struct {
	Async		bool
	Bootstrapsvrs	string
	Clientid	string
	CidUseHostname	bool
	ClientOpts	[]GenericKV
	FlushWait	int
}

type pubsubIntfKafka struct {
	cfg		PubsubIntfConfDrvKafka
	debug		bool
	hostname	string

	kafkaProducer	*kafka.Producer
}

func pubsubIntfKafkaNew(cfg PubsubIntfConf) (*pubsubIntfKafka, error) {
	var err error 

	// start building interface
	r := &pubsubIntfKafka {
		cfg:		cfg.DriverKafka,
		debug:		cfg.Debug,
	}

	// client id
	r.hostname = r.cfg.Clientid
	if r.cfg.CidUseHostname {
		r.hostname, err = os.Hostname()
		if err != nil {
			log.Printf("client_id_use_hostname is true but could not get hostname: %v", err)
			return nil, err
		}
	} else if r.hostname == "" {
		err = fmt.Errorf("client_id_use_hostname is false but client_id is not specified")
		return nil, err
	} 

	// ready
	err = r.initKafkaClient()
	if err != nil {
		return nil, err
	}
	log.Printf("Kafka client ready: %v", r.kafkaProducer)

	return r, nil
}

func (s *pubsubIntfKafka) buildKafkaClientConf() kafka.ConfigMap {
	r := make(map[string]kafka.ConfigValue)
	r["bootstrap.servers"] = s.cfg.Bootstrapsvrs
	r["client.id"] = s.hostname
	for _, v := range(s.cfg.ClientOpts) {
		r[v.Key] = v.Value
	}

	return r
}

func (s *pubsubIntfKafka) initKafkaClient() error {
	var err error

	// init producer
	c := s.buildKafkaClientConf()
	s.kafkaProducer, err = kafka.NewProducer(&c)
	if err != nil {
		return err
	}

	// init event handler if async mode
	if s.cfg.Async {
		go func() {
			for e := range s.kafkaProducer.Events() {
				s.handleKafkaEvent(e)
			}
		}()
	}

	return nil
}

func (s *pubsubIntfKafka) handleKafkaEvent(e kafka.Event) {
	switch ev := e.(type) {
		case *kafka.Message:
			m := ev
			if m.TopicPartition.Error != nil {
				log.Printf("Kafka Event: Failed to deliver message (%v)", 
					m.TopicPartition.Error)
			} else if s.debug {
				log.Printf("Kafka Event: Delivered message to topic %s [%d] at offset %v",
					*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
			}
		case kafka.Error:
			log.Printf("Kafka Event: generic client error occured (%v)", ev)
		default:
			log.Printf("Kafka Event: unhandled Kafka event (%v)", ev)

	}

	return
}

func (s *pubsubIntfKafka) Close() error {
	t := 0
	for s.kafkaProducer.Flush(1000) > 0 {
		if t > s.cfg.FlushWait {
			log.Printf("Giving up on flush after 30 seconds..")
			break
		}
		log.Printf("Waiting for Kafka client to flush outstanding messages..")
		t++
	}

	s.kafkaProducer.Close()
	
	return nil
}

func (s *pubsubIntfKafka) PublishRaw(target PubsubIntfTarget, data string) error {
	if s.debug {
		log.Printf("Publish data %s to target %v", data, target)
	}

	// build message
	var msgHeader []kafka.Header
	for _, v := range(target.Header) {
		hdrEntry := kafka.Header {
			Key: v.Key,
			Value: []byte(v.Value),
		}
		msgHeader = append(msgHeader, hdrEntry)
	}
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &target.Topic, Partition: kafka.PartitionAny},
		Value:          []byte(data),
		Headers:        msgHeader,
	}

	// send
	if s.cfg.Async {
		err := s.kafkaProducer.Produce(msg, nil)
		if err != nil {
			log.Printf("Failed to publish message: %v", err)
			return err
		}
	} else {
		delivery_chan := make(chan kafka.Event, 10000)
		err := s.kafkaProducer.Produce(msg, delivery_chan)
		
		e := <-delivery_chan
		s.handleKafkaEvent(e)

		if err != nil { 
			log.Printf("Failed to publish message: %v", err)
			return err
		}
	}

	return nil
}
