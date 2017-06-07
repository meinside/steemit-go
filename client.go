package steemit

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// https://steemit.com/witness-category/@bitcoiner/list-of-public-steem-full-api-nodes-and-example-usage
	ApiAddr = "wss://steemd.steemit.com"

	TimeoutSeconds = 5
	PingSeconds    = 30
)

type Request struct {
	Id     int64         `json:"id"`
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

type Response struct {
	Id     int64       `json:"id"`
	Result interface{} `json:"result,omitempty"`
	Error  struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Code    int    `json:"code"`
			Name    string `json:"name"`
			Message string `json:"message"`
			Stack   []struct {
				Context struct {
					Level      string `json:"level"`
					File       string `json:"file"`
					Line       int    `json:"line"`
					Method     string `json:"method"`
					Hostname   string `json:"hostname"`
					ThreadName string `json:"thread_name"`
					Timestamp  string `json:"timestamp"` // format: "2017-06-07T02:28:07"
				} `json:"context"`
				Format string      `json:"format"`
				Data   interface{} `json:"data"`
			} `json:"stack"`
		} `json:"data"`
	} `json:"error,omitempty"`
}

// Client
type Client struct {
	conn      *websocket.Conn
	callbacks map[int64]func(response Response, err error)
	lastId    int64
	isVerbose bool

	sync.Mutex
}

// Create a new client
func NewClient(verbose bool) (*Client, error) {
	// connect,
	conn, _, err := websocket.DefaultDialer.Dial(ApiAddr, nil)
	if err != nil {
		log.Fatal("Failed to dial:", err)

		return nil, err
	}

	newClient := &Client{
		conn:      conn,
		callbacks: make(map[int64]func(Response, error)),
		lastId:    0,
		isVerbose: verbose,
	}

	// wait for incoming messages,
	go func(client *Client) {
		for {
			if typ, bytes, err := client.conn.ReadMessage(); err != nil {
				log.Println("Failed to read:", err)
			} else {
				if typ == websocket.TextMessage {
					if client.isVerbose {
						log.Printf("Received bytes: %s", string(bytes))
					}

					var res Response
					if err := json.Unmarshal(bytes, &res); err == nil {
						if fn, exists := client.callbacks[res.Id]; exists {
							go fn(res, nil) // call callback function

							delete(client.callbacks, res.Id) // delete used callback function
						} else {
							log.Printf("No such callback with id: %d", res.Id)
						}
					} else {
						log.Printf("Failed to decode received message: %s", err)
					}
				}
			}
		}
	}(newClient)

	// send ping (for keeping connection)
	go func(client *Client) {
		for {
			time.Sleep(PingSeconds * time.Second)

			client.Lock()
			err := client.conn.WriteMessage(websocket.PingMessage, []byte{})
			client.Unlock()

			if err != nil {
				log.Printf("Failed to ping: %s", err)
			}
		}
	}(newClient)

	return newClient, nil
}

// Get a new request
func (c *Client) NewRequest(api, function string, params []interface{}) Request {
	req := Request{
		Method: "call",
		Params: []interface{}{
			api,
			function,
			params,
		},
		Id: c.lastId,
	}
	c.lastId += 1 // increase last id

	return req
}

// Send request and return its result through callback function
func (c *Client) SendRequestAsync(
	request Request,
	callback func(r Response, e error),
) (err error) {
	var bytes []byte
	if bytes, err = json.Marshal(request); err == nil {
		c.callbacks[request.Id] = callback

		c.Lock()
		err = c.conn.WriteMessage(websocket.TextMessage, bytes)
		c.Unlock()
	}

	return err
}

// Send request and return its result synchronously
func (c *Client) SendRequest(request Request) (r Response, err error) {
	rCh := make(chan Response)
	eCh := make(chan error)

	err = c.SendRequestAsync(
		request,
		func(r Response, e error) {
			if e != nil {
				eCh <- e
			} else {
				rCh <- r
			}
		},
	)
	if err == nil {
		// wait for the callback
		select {
		case r := <-rCh:
			return r, nil
		case e := <-eCh:
			err = e
		case <-time.After(time.Second * TimeoutSeconds): // timeout
			err = fmt.Errorf("Request timed out in %d sec", TimeoutSeconds)
		}
	}

	return Response{}, err
}

// Close connection
func (c *Client) Close() error {
	c.Lock()
	defer c.Unlock()

	return c.conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	)
}
