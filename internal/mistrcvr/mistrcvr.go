package mistrcvr

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"mist-to-tsdb/internal/wsclient"
	"mist-to-tsdb/internal/tsdb"
)

type Rcvr struct {
	cfg	Config

	client	*wsclient.WsClient
	tsdb	*tsdb.TsdbIntf
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
		BufSize:	cfg.Mist.BufSize,
		Subscriptions:	subs,
	}

	r.client, err = wsclient.New(clientConf)
	if err != nil {
		return nil, err
	}

	// TSDB Client Initialization
	tsdbConf := tsdb.TsdbIntfConf {
		Debug:		cfg.Tsdb.Debug,
		Driver:		cfg.Tsdb.Driver,
		Datasource:	cfg.Datasource,
		DataInChannel:	r.client.GetDataChannel(),
	}
	tsdbConf.DriverAwsTimeStream.Region = cfg.Tsdb.Awstimestream.Region
	tsdbConf.DriverAwsTimeStream.Database = cfg.Tsdb.Awstimestream.Database
	tsdbConf.DriverAwsTimeStream.MaxRetries = cfg.Tsdb.Awstimestream.Maxretries
	r.tsdb, err = tsdb.New(tsdbConf)
	if err != nil {
		return nil, err
	}

	return r, nil 
}

func (r *Rcvr) Run() error {
	// Launch
	clientShutdownSig := make(chan struct{}, 1)
	go r.client.Run(r.wg, clientShutdownSig)

	tsdbShutdownSig := make(chan struct{}, 1)
	go r.tsdb.Run(r.wg, tsdbShutdownSig)

	// Main thread to wait until we get a kill signal or something go wrong
	killSig := make(chan os.Signal, 1)
	signal.Notify(killSig, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	<-killSig

	log.Printf("Caught kill signal, shutting down")
	close(clientShutdownSig)
	close(tsdbShutdownSig)
	r.wg.Wait()

	log.Printf("All threads exited")

	return nil
}
