package mqtt_client

import (
	"fmt"
	"log"
	"sync"

	"github.com/alext/heating-controller/config"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Client struct {
	lock          sync.RWMutex
	subscriptions map[string][]chan string
	client        mqtt.Client
}

func New(cfg *config.MQTTConfig) (*Client, error) {

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", cfg.Host, cfg.Port))
	opts.SetClientID("heating-controller")
	opts.SetOrderMatters(false)
	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
		opts.SetPassword(cfg.Password)
	}
	cl := mqtt.NewClient(opts)
	if token := cl.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return &Client{
		subscriptions: make(map[string][]chan string),
		client:        cl,
	}, nil
}

func (c *Client) topicSubscribe(topic string) error {
	var qos byte = 0

	token := c.client.Subscribe(topic, qos, c.handleMessage)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (c *Client) Subscribe(topic string) (<-chan string, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	subs := c.subscriptions[topic]
	if len(subs) == 0 {
		err := c.topicSubscribe(topic)
		if err != nil {
			return nil, err
		}
	}

	ch := make(chan string, 1)
	c.subscriptions[topic] = append(subs, ch)
	return ch, nil
}

func (c *Client) handleMessage(_ mqtt.Client, msg mqtt.Message) {
	c.lock.RLock()
	subs, ok := c.subscriptions[msg.Topic()]
	c.lock.RUnlock()

	if !ok {
		log.Printf("[mqtt] ignoring message for unexpected topic %s", msg.Topic())
		return
	}

	payload := string(msg.Payload())
	for _, ch := range subs {
		select {
		case ch <- payload:
		default:
		}
	}
}
