package tsdb

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"mist-to-tsdb/pkg/mistdatafmt"
)

type tsdbIntfBackend interface {
	AddRecordStatsClient(string, mistdatafmt.WsMsgClientStat) error
	AddRecordRaw(string, string) error
}

type TsdbIntfConf struct {
	Debug			bool		`mapstructure:"debug",default:false`
	Driver			string		`mapstructure:"driver",default:"awstimestream"`
	DriverAwsTimeStream	struct {
		Region		string		`mapstructure:"aws_region",default:"us-east-1"`
		Database	string		`mapstructure:"database"`
		MaxRetries	int		`mapstructure:"max_retries",default:3`
	}                                       `mapstructure:"aws_timestream"`

	Datasource		[]TsdbIntfConfDS
	DataInChannel		chan mistdatafmt.WsMsgData
}

type TsdbIntfConfDS struct {
	Channel		string	  `mapstructure:"channel"`
	Datalayout	string	  `mapstructure:"data_layout"`
	Table		string	  `mapstructure:"table"`
	Keys		[]string  `mapstructure:"keys"`
	Metrics []struct {
		Key	string	  `mapstructure:"key"`
		Type	string	  `mapstructure:"type"`
	}                         `mapstructure:"metrics"`
}

type TsdbIntf struct {
	cfg		TsdbIntfConf
	backend		tsdbIntfBackend
	dataIn		chan mistdatafmt.WsMsgData
	layoutMap	map[string]int
	wg		*sync.WaitGroup
}

const (
	LAYOUT_STATS_CLIENT = iota
	LAYOUT_RAW
	LAYOUT_NULL
)

func New(cfg TsdbIntfConf) (*TsdbIntf, error) {
	var err error
	r := &TsdbIntf {
		cfg:		cfg,
		dataIn:		cfg.DataInChannel,
		layoutMap:	make(map[string]int),
	}

	// Populate layout mapping
	for i := 0; i < len(cfg.Datasource); i++ {
		l := strings.ToLower(cfg.Datasource[i].Datalayout)
		switch l {
		case "stats_client":
			r.layoutMap[cfg.Datasource[i].Channel] = LAYOUT_STATS_CLIENT
		case "raw":
			r.layoutMap[cfg.Datasource[i].Channel] = LAYOUT_RAW
		default:
			return nil, fmt.Errorf("Unsupported layout %s", l)
		}
	}

	// Init Backend Driver
	switch cfg.Driver {
	case "awstimestream":
		r.backend, err = tsdbIntfAwsTsNew(cfg)
		if err != nil {
			return nil, err
		}

	case "dummy":
		r.backend, err = tsdbIntfDummyNew(cfg)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("Unknown TSDB driver: %s", cfg.Driver)
	}

	return r, nil
}

func (i *TsdbIntf) Run(wg *sync.WaitGroup, killSig chan struct{}) error {
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
				log.Printf("TSDB Start Process: %v", msg)
			}

			err = i.processData(msg.Channel, msg.Data)
			if err != nil {
				log.Printf("TSDB driver has thrown error: %v", err)
				log.Printf("Failed data: %v", msg)
			}
		}
	}

	return nil
}

func (i *TsdbIntf) processData(channel string, data string) error {
	var err error

	layout, e := i.layoutMap[channel]
	if !e {
		return fmt.Errorf("Data received on channel %s, but no layout defined", channel)
	}
		
	switch layout {
	case LAYOUT_STATS_CLIENT:
		jsonData := &mistdatafmt.WsMsgClientStat{}
		err = json.Unmarshal([]byte(data), jsonData)
		if err != nil {
			return err
		}

		err = i.backend.AddRecordStatsClient(channel, *jsonData)
		if err != nil {
			return err
		}

	case LAYOUT_RAW:
		err = i.backend.AddRecordRaw(channel, data)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("Unsupported data layout for channel %s", channel)
	}
	
	return nil
}

func (i *TsdbIntf) finish() {
	if i.wg != nil {
		i.wg.Done()
	}

	return
}
