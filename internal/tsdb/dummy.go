package tsdb

import (
	"fmt"
	"log"

	"mist-to-tsdb/pkg/mistdatafmt"
)

type tsdbIntfDummy struct {
	dataOut		map[string]bool
}



func tsdbIntfDummyNew(cfg TsdbIntfConf) (*tsdbIntfDummy, error) {
	// start building interface
	r := &tsdbIntfDummy {
		dataOut:	make(map[string]bool),
	}

	// build data out
	for i := 0; i < len(cfg.Datasource); i++ {
		if cfg.Datasource[i].Channel == "" {
			return nil, fmt.Errorf("Missing mandatory data out parameter")
		}

		r.dataOut[cfg.Datasource[i].Channel] = true
	}

	return r, nil
}

func (i *tsdbIntfDummy) AddRecordStatsClient(channel string, data mistdatafmt.WsMsgClientStat) error {
	_, ok := i.dataOut[channel]
	if !ok {
		return fmt.Errorf("Data out parameters for channel %s was not found", channel)
	}

	log.Printf("[TSDB] AddRecordStatsClient Channel %s Data %v", channel, data)	

	return nil
}

func (i *tsdbIntfDummy) AddRecordRaw(channel string, data string) error {
	_, ok := i.dataOut[channel]
	if !ok {
		return fmt.Errorf("Data out parameters for channel %s was not found", channel)
	}

	log.Printf("[TSDB] AddRecordRaw Channel %s Data %s", channel, data)	

	return nil
}

