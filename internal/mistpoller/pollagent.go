package mistpoller

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/yumyudai/misttools/internal/common"
	"github.com/yumyudai/misttools/pkg/mistdatafmt"
)

type PollAgent struct {
	Id		int
	Endpoint	string
	Apikey		string
	Uri		string
	Layout		string
	Interval	int
	WatchKeys	[]string
	UniqueKey	string
	Out		chan common.MistApiData
	Debug		bool

	intvlTicker	*time.Ticker
	prevData	map[string]mistdatafmt.MistDataFmtIntf
	killSig		chan struct{}
	wg		*sync.WaitGroup
}

func (s *PollAgent) Run(wg *sync.WaitGroup, killSig chan struct{}) error {
	log.Printf("agent#%d: start poll agent thread (uri %s, interval %d)", s.Id, s.Uri, s.Interval)

	// init
	s.intvlTicker = time.NewTicker(time.Duration(s.Interval) * time.Second)
	s.prevData = make(map[string]mistdatafmt.MistDataFmtIntf)
	s.killSig = killSig
	s.wg = wg

	// start
	wg.Add(1)
	defer s.finish()

	s.runRequest()
	for {
		select {
		case <-killSig:
			return nil
		case <-s.intvlTicker.C:
			s.runRequest()
		}
	}

	return nil
}

func (s *PollAgent) CheckParams() error {
	switch(s.Layout) {
	case "maps":
		if s.UniqueKey == "" {
			return fmt.Errorf("agent#%d: unique is not set which is mandatory for array based response", s.Id)
		}

	case "zones":
		if s.UniqueKey == "" {
			return fmt.Errorf("agent#%d: unique is not set which is mandatory for array based response", s.Id)
		}

	case "raw":
		break

	default:
		return fmt.Errorf("agent#%d: unknown data layout %s", s.Id, s.Layout)
	}

	return nil
}

func (s *PollAgent) finish() {
	if s.intvlTicker != nil {
		s.intvlTicker.Stop()
	}

	if s.wg != nil {
		s.wg.Done()
	}

	log.Printf("agent#%d: finished process thread", s.Id)

	return
}

func (s *PollAgent) runRequest() {
	// build request
	reqUrl := "" 
	if !strings.HasPrefix(s.Endpoint, "http://") && !strings.HasPrefix(s.Endpoint, "https://") {
		reqUrl = "https://" + s.Endpoint + s.Uri
	} else {
		reqUrl = s.Endpoint + s.Uri
	}

	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		log.Printf("agent#%d: failed to build HTTP request (%v)", s.Id, err)
		return
	}

	// set authentication header
	tokenStr := fmt.Sprintf("token %s", s.Apikey)
	req.Header.Set("Authorization", tokenStr)

	// start requet
	if s.Debug {
		log.Printf("agent#%d: start HTTP GET request: url %s", s.Id, reqUrl)
	}
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("agent#%d: HTTP request failure (%v)", s.Id, err)
		return
	}
	defer resp.Body.Close()

	// read response
	if s.Debug {
		log.Printf("agent#%d: got HTTP response: %-v", s.Id, resp)
	}

	if resp.StatusCode != 200 {
		log.Printf("agent#%d: HTTP request has returned status code %d", s.Id, resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("agent#%d: Failed to read HTTP response body (%v)", s.Id, err)
		return
	}

	s.processData(string(body))
	return
}

func (s *PollAgent) processData(data string) {
	switch(s.Layout) {
	case "maps":
		entries := make([]*mistdatafmt.ApiDataMapEntry, 0)
		err := json.Unmarshal([]byte(data), &entries)
		if err != nil {
			log.Printf("agent#%d: failed to parse JSON (%v)", s.Id, err)
			return
		}

		if s.Debug {
			log.Printf("agent#%d: got %d entries", s.Id, len(entries))
		}

		if s.UniqueKey == "" {
			log.Printf("agent#%d: unique is not set which is mandatory for array based response", s.Id)
			return
		}

		needPublish := false
		currData := make(map[string]mistdatafmt.MistDataFmtIntf)
		for _, entry := range(entries) {
			uniqueKeyVal, err := entry.GetJsonKeyValueAsStr(s.UniqueKey)
			if err != nil {
				log.Printf("agent%d: failed to get value of unique key %s (%v)", s.Id, s.UniqueKey, err)
				continue
			}

			prev, exists := s.prevData[uniqueKeyVal]
			if !exists || s.dataHasChanged(prev, entry) {
				if s.Debug {
					log.Printf("agent#%d: difference in key %s exists=%v", s.Id, uniqueKeyVal, exists)
				}
				needPublish = true
			}
			currData[uniqueKeyVal] = entry
		}

		for key, _ := range(s.prevData) {
			_, ok := currData[key]
			if !ok {
				log.Printf("agent#%d: key has been deleted %s", s.Id, key)
				needPublish = true
			}
		}

		// sent all keys so that delete can be handled on receiver
		if needPublish {
			log.Printf("agent#%d: publish data %s", s.Id, data)
			s.doPublish(data)
			s.prevData = currData
		} else {
			log.Printf("agent%d: no need to publish data")
		}

	case "zones":
		entries := make([]*mistdatafmt.ApiDataZoneEntry, 0)
		err := json.Unmarshal([]byte(data), &entries)
		if err != nil {
			log.Printf("agent#%d: failed to parse JSON (%v)", s.Id, err)
			return
		}

		if s.Debug {
			log.Printf("agent#%d: got %d entries", s.Id, len(entries))
		}

		if s.UniqueKey == "" {
			log.Printf("agent#%d: unique is not set which is mandatory for array based response", s.Id)
			return
		}

		needPublish := false
		currData := make(map[string]mistdatafmt.MistDataFmtIntf)
		for _, entry := range(entries) {
			uniqueKeyVal, err := entry.GetJsonKeyValueAsStr(s.UniqueKey)
			if err != nil {
				log.Printf("agent%d: failed to get value of unique key %s (%v)", s.Id, s.UniqueKey, err)
				continue
			}

			// TODO: zone vertice may need to be sorted..
			prev, exists := s.prevData[uniqueKeyVal]
			if !exists || s.dataHasChanged(prev, entry) {
				if s.Debug {
					log.Printf("agent#%d: difference in key %s exists=%v", s.Id, uniqueKeyVal, exists)
				}
				needPublish = true
			}
			currData[uniqueKeyVal] = entry
		}

		for key, _ := range(s.prevData) {
			_, ok := currData[key]
			if !ok {
				log.Printf("agent#%d: key has been deleted %s", s.Id, key)
				needPublish = true
			}
		}

		// sent all keys so that delete can be handled on receiver
		if needPublish {
			log.Printf("agent#%d: publish data", s.Id, data)
			s.doPublish(data)
			s.prevData = currData
		} else {
			log.Printf("agent%d: no need to publish data", s.Id)
		}

	case "raw":
		if s.Debug {
			log.Printf("agent#%d: publish raw")
		}
		s.doPublish(data)

	default:
		log.Printf("agent#%d: unknown data layout %s", s.Id, s.Layout)
		return
	}

	return
}

func (s *PollAgent) dataHasChanged(prevData mistdatafmt.MistDataFmtIntf, currData mistdatafmt.MistDataFmtIntf) bool {
	for _, key := range(s.WatchKeys) {
		uKeyVal, _ := prevData.GetJsonKeyValueAsStr(s.UniqueKey) // for debugging

		prevVal, err := prevData.GetJsonKeyValue(key)
		if err != nil {
			log.Printf("agent#%d: failed to get json key %s (%v)", s.Id, key, err)
			return false
		}

		currVal, err := currData.GetJsonKeyValue(key)
		if err != nil {
			log.Printf("agent#%d: failed to get json key %s (%v)", s.Id, key, err)
			return false
		}

		if !reflect.DeepEqual(prevVal, currVal) {
			if s.Debug {
				log.Printf("agent#%d: %s has changed (key %s prev %s curr %s)", s.Id, uKeyVal, key, prevVal, currVal)
			} 
			return true
		} else if s.Debug {
			log.Printf("agent#%d: %s data has not changed (%s)", s.Id, uKeyVal, currVal)
		}
	}

	return false
}
func (s *PollAgent) doPublish(data string) {
	out := common.MistApiData {
		Origin: s.Uri,
		Data: data,
	}

	s.Out <-out
	return
}
