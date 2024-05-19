package mistwsrcvr

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"mist-to-tsdb/internal/wsclient"
	"mist-to-tsdb/internal/tsdb"
	"mist-to-tsdb/internal/pubsub"
	"mist-to-tsdb/pkg/mistdatafmt"
)

type Rcvr struct {
	cfg	Config

	client	*wsclient.WsClient
	tsdb	*tsdb.TsdbIntf
	pubsub	*pubsub.PubsubIntf
	wg	*sync.WaitGroup
}

func New(cfg Config) (*Rcvr, error) {
	var err error

	// Base Initialization
	r := &Rcvr {
		cfg:	cfg,
		wg:	&sync.WaitGroup{},
	}

	// Mist WebSocket Client Initialization
	var subs []string
	for _, v := range(cfg.Datasource) {
		subs = append(subs, v.Channel)
	}

	clientConf := wsclient.WsClientConf {
		ApiEndpoint:	cfg.Mist.Endpoint,
		ApiKey:		cfg.Mist.Apikey,
		Debug:		cfg.Mist.Debug,
		Subscriptions:	subs,
	}

	r.client, err = wsclient.New(clientConf)
	if err != nil {
		return nil, err
	}

	// TSDB Client Initialization
	if cfg.Tsdb.Enabled {
		var tsdbDSs []tsdb.TsdbIntfConfDS
		for _, v := range(cfg.Datasource) {
			ds := tsdb.TsdbIntfConfDS {
				Channel:	v.Channel,
				Datalayout:	v.Datalayout,
				Table:		v.Tsdb.Table,
				Keys:		v.Tsdb.Keys,
				Metrics:	v.Tsdb.Metrics,
			}

			tsdbDSs = append(tsdbDSs, ds)
		}

		tsdbChan := make(chan mistdatafmt.WsMsgData, cfg.Tsdb.BufSize)
		err = r.client.AddDataChannel(tsdbChan)
		if err != nil {
			return nil, err
		}

		tsdbConf := tsdb.TsdbIntfConf {
			Debug:		cfg.Tsdb.Debug,
			Driver:		cfg.Tsdb.Driver,
			Datasource:	tsdbDSs,
			DataInChannel:	tsdbChan,
		}
		tsdbConf.DriverAwsTimeStream.Region = cfg.Tsdb.Awstimestream.Region
		tsdbConf.DriverAwsTimeStream.Database = cfg.Tsdb.Awstimestream.Database
		tsdbConf.DriverAwsTimeStream.MaxRetries = cfg.Tsdb.Awstimestream.Maxretries
		r.tsdb, err = tsdb.New(tsdbConf)
		if err != nil {
			return nil, err
		}
	}

	// PubSub Client Initialization
	if cfg.Pubsub.Enabled {
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
				Channel:	v.Channel,
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


		pubsubChan := make(chan mistdatafmt.WsMsgData, cfg.Pubsub.BufSize)
		err = r.client.AddDataChannel(pubsubChan)
		if err != nil {
			return nil, err
		}

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
	}

	return r, nil 
}

func (r *Rcvr) Run() error {
	var shutdownSigs []chan struct{}
	// Launch
	clientShutdownSig := make(chan struct{}, 1)
	shutdownSigs = append(shutdownSigs, clientShutdownSig)
	go r.client.Run(r.wg, clientShutdownSig)

	if r.cfg.Tsdb.Enabled {
		tsdbShutdownSig := make(chan struct{}, 1)
		shutdownSigs = append(shutdownSigs, tsdbShutdownSig)
		go r.tsdb.Run(r.wg, tsdbShutdownSig)
	}

	if r.cfg.Pubsub.Enabled {
		pubsubShutdownSig := make(chan struct{}, 1)
		shutdownSigs = append(shutdownSigs, pubsubShutdownSig)
		go r.pubsub.Run(r.wg, pubsubShutdownSig)
	}

	// Main thread to wait until we get a kill signal or something go wrong
	killSig := make(chan os.Signal, 1)
	signal.Notify(killSig, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	<-killSig

	log.Printf("Caught kill signal, shutting down")
	for _, sig := range(shutdownSigs) {
		close(sig)
	}
	r.wg.Wait()

	log.Printf("All threads exited")

	return nil
}
