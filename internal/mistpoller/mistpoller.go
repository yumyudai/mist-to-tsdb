package mistpoller

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/yumyudai/misttools/internal/common"
	"github.com/yumyudai/misttools/internal/pubsub"
)

type Poller struct {
	cfg	Config

	agents	[]*PollAgent
	pubsub	*pubsub.PubsubIntf
	wg	*sync.WaitGroup
}

func New(cfg Config) (*Poller, error) {
	var err error

	// Base Initialization
	r := &Poller {
		cfg:	cfg,
		agents:	make([]*PollAgent, 0),
		wg:	&sync.WaitGroup{},
	}

	// PubSub Client Initialization
	var pubsubDSs []pubsub.PubsubIntfTarget
	for _, v := range(cfg.Datasource) {
		var hdr []pubsub.GenericKV
		for _, v := range(v.Pubsub.Header) {
			e := pubsub.GenericKV {
				Key: v.Key,
				Value: v.Value,
			}

			hdr = append(hdr, e)
		}

		ds := pubsub.PubsubIntfTarget {
			Channel:	v.Uri,
			Topic:		v.Pubsub.Topic,
			Header:		hdr,
		}

		pubsubDSs = append(pubsubDSs, ds)
	}

	var kafkaClientOpts []pubsub.GenericKV
	for _, v := range(cfg.Pubsub.Kafka.ClientOpts) {
		opt := pubsub.GenericKV {
			Key:	v.Key,
			Value:	v.Value,
		}

		kafkaClientOpts = append(kafkaClientOpts, opt)
	}
	pubsubKafkaConf := pubsub.PubsubIntfConfDrvKafka {
		Async:		cfg.Pubsub.Kafka.Async,
		Bootstrapsvrs:	cfg.Pubsub.Kafka.Bootstrapsvrs,
		Clientid:	cfg.Pubsub.Kafka.Clientid,
		CidUseHostname:	cfg.Pubsub.Kafka.CidUseHostname,
		ClientOpts:	kafkaClientOpts,
		FlushWait:	cfg.Pubsub.Kafka.FlushWait,
	}

	pubsubChan := make(chan common.MistApiData, cfg.Pubsub.BufSize)
	pubsubConf := pubsub.PubsubIntfConf {
		Debug:		cfg.Pubsub.Debug,
		Driver:		cfg.Pubsub.Driver,
		DriverKafka:	pubsubKafkaConf,
		Datasource:	pubsubDSs,
		DataInChannel:	pubsubChan,
	}
	r.pubsub, err = pubsub.New(pubsubConf)
	if err != nil {
		return nil, err
	}

	// Poll Agent Initialization
	for id, v := range(cfg.Datasource) {
		agent := &PollAgent {
			Id:		id,
			Endpoint:	cfg.Mist.Endpoint,
			Apikey:		cfg.Mist.Apikey,
			Uri:		v.Uri,
			Layout:		v.Datalayout,
			Interval:	v.Interval,
			WatchKeys:	v.WatchKeys,
			UniqueKey:	v.UniqueKey,
			Out:		pubsubChan,
			Debug:		cfg.Mist.Debug,
		}

		err = agent.CheckParams()
		if err != nil {
			return nil, err
		}
	
		r.agents = append(r.agents, agent)
		id++
	}

	return r, nil 
}

func (s *Poller) Run() error {
	var shutdownSigs []chan struct{}
	// Launch
	for _, agent := range(s.agents) {
		agentShutdownSig := make(chan struct{}, 1)
		shutdownSigs = append(shutdownSigs, agentShutdownSig)
		go agent.Run(s.wg, agentShutdownSig)
	}

	pubsubShutdownSig := make(chan struct{}, 1)
	shutdownSigs = append(shutdownSigs, pubsubShutdownSig)
	go s.pubsub.Run(s.wg, pubsubShutdownSig)

	// Main thread to wait until we get a kill signal or something go wrong
	killSig := make(chan os.Signal, 1)
	signal.Notify(killSig, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	<-killSig

	log.Printf("Caught kill signal, shutting down")
	for _, sig := range(shutdownSigs) {
		close(sig)
	}
	s.wg.Wait()

	log.Printf("All threads exited")

	return nil
}
