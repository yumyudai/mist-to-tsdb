package tsdb

import (
	"log"

	"mist-to-tsdb/pkg/mistdatafmt"
)

type tsdbIntfDummy struct{}

func tsdbIntfDummyNew(cfg TsdbIntfConf) (*tsdbIntfDummy, error) {
	r := &tsdbIntfDummy{}
	return r, nil
}

func (i *tsdbIntfDummy) AddRecordStatsClient(channel string, data mistdatafmt.WsMsgClientStat) error {
	log.Printf("[TSDB] AddRecordStatsClient Channel %s Data %v", channel, data)	
	return nil
}

func (i *tsdbIntfDummy) AddRecordRaw(channel string, data string) error {
	log.Printf("[TSDB] AddRecordRaw Channel %s Data %s", channel, data)	
	return nil
}

