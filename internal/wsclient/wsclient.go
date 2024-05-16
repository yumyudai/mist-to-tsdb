package wsclient

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"mist-to-tsdb/pkg/mistdatafmt"
)

type WsClientConf struct {
	ApiEndpoint	string
	ApiKey		string
	Debug		bool	`default:false`

	Subscriptions	[]string
}

type WsClient struct {
	cfg		WsClientConf
	endpoint	url.URL
	msgChans	[]chan mistdatafmt.WsMsgData
	wsConn		*websocket.Conn
	wg		*sync.WaitGroup
} 

var (
	ErrShutdown	= fmt.Errorf("Shutdown")
)

func New(cfg WsClientConf) (*WsClient, error) {
	// Check Input
	if cfg.ApiKey == "" || cfg.ApiEndpoint == "" {
		return nil, fmt.Errorf("Required parameters were not given")
	}
	if len(cfg.Subscriptions) < 1 {
		return nil, fmt.Errorf("No data subscriptions specified")
	}

	// Build Client
	r := &WsClient {
		cfg:		cfg,
		endpoint:	url.URL {
					Scheme:	"wss",
					Host:	cfg.ApiEndpoint,
					Path:	"/api-ws/v1/stream",
				},
		wsConn:		nil,
		wg:		nil,
	}

	if cfg.Debug {
		log.Printf("WebSocket client debugging is ON")
	}

	return r, nil
}

func (c *WsClient) AddDataChannel(newChan chan mistdatafmt.WsMsgData) error {
	c.msgChans = append(c.msgChans, newChan)
	return nil
}

func (c *WsClient) Run(wg *sync.WaitGroup, killSig chan struct{}) error {
	var err error

	// Final pre-flight init
	c.wg = wg
	wg.Add(1)
	defer c.finish()

	// Launch
	for {
		err = c.initConn()
		if err != nil {
			log.Printf("Failed to connect: %v", err)
		} else {
			err = c.readLoop(killSig)
			if err == ErrShutdown {
				log.Printf("Shutting down WebSocket client..")
				break
			} else if err != nil {
				log.Printf("Read loop has exited abnormaly: %v", err)
			}
		}

		log.Printf("Reconnect after 10 seconds..")
		time.Sleep(10 * time.Second)
	}

	return nil
}

func (c *WsClient) initConn() error {
	var err error

	log.Printf("Connecting to WebSocket endpoint: %s", c.endpoint.String())
	
	// Preparation
	tokenStr := fmt.Sprintf("token %s", c.cfg.ApiKey)
	authHeader := http.Header{}
	authHeader.Set("Authorization", tokenStr)

	// Connect
	c.wsConn, _, err = websocket.DefaultDialer.Dial(c.endpoint.String(), authHeader)
	if err != nil {
		return fmt.Errorf("Failed to dial: %v", err)
	}

	// Send Subscription Requests
	for i := 0; i < len(c.cfg.Subscriptions); i++ {
		err = c.sendSubscribe(c.cfg.Subscriptions[i])
		if err != nil {
			// continue
			log.Printf("%v", err)
		}
	}

	return nil
}

func (c *WsClient) finish() {
	if c.wsConn != nil {
		c.wsConn.Close()
	}

	if c.wg != nil {
		c.wg.Done()
	}

	return
}

func (c *WsClient) readLoop(killChan chan struct{}) error {
	dataChan := make(chan *mistdatafmt.WsMsgData, 1)
	go func() {
		for {
			msgType, data, err := c.wsConn.ReadMessage()
			if err != nil {
				log.Printf("Failed to read message: %v", err)
				close (dataChan)
				return
			}

			if c.cfg.Debug {
				log.Printf("Got message (type: %d): %s", msgType, data)
			}

			if msgType != websocket.TextMessage {
				log.Printf("Got message type that is not text")
				continue
			}

			wsmsg := &mistdatafmt.WsMsgData{}
			err = json.Unmarshal(data, wsmsg)
			if err != nil {
				log.Printf("Failed to read JSON: %v", err)
				continue
			}

			dataChan <-wsmsg
		}
	}()

	for {
		select {
		case <-killChan:
			return ErrShutdown
		case wsmsg :=<-dataChan:
			if wsmsg == nil {
				return fmt.Errorf("Data channel has closed..")
			}
			c.processMsg(wsmsg)
		}
	}
	
	return nil
}

func (c *WsClient) sendSubscribe(channel string) error {
	var err error

	// Build Message
	req := mistdatafmt.WsMsgSubscribe {
		Subscribe:	channel,
	}
	
	// Send
	log.Printf("here")
	err = c.wsConn.WriteJSON(req)
	if err != nil {
	log.Printf("here3")
		return fmt.Errorf("Failed to send message: %v", err)
	}
	log.Printf("here2")

	return nil
}

func (c *WsClient) processMsg(m *mistdatafmt.WsMsgData) {
	switch m.Event {
	case "channel_subscribed":
		log.Printf("Subscription Successful: %s", m.Channel)

	case "subscribe_failed":
		log.Printf("Subscription Failed: %s (%s)", m.Channel, m.Detail)

	case "data":
		if c.cfg.Debug {
			log.Printf("Recv WebSocket Data: Channel %s Data %s", m.Channel, m.Data)
		}

		for _, ch := range(c.msgChans) {
			ch <-*m
		}

	default:
		log.Printf("Received event not implemented: %s", m.Event)
	}

	return
}
