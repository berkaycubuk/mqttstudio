package main

import (
	"fmt"
	"log"
	"os"
	"slices"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Connection struct {
	ProjectID int
	Status int // 0: offline, 1: online
	Broker string
	ClientID string
	Client mqtt.Client
	Topics []string
	DataBuffer map[string][][]byte
}

func init() {
	mqtt.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
	mqtt.CRITICAL = log.New(os.Stdout, "[CRIT] ", 0)
	mqtt.WARN = log.New(os.Stdout, "[WARN]  ", 0)
	mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)
}

func (c *Connection) Connect() error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(c.Broker)
	opts.SetClientID(c.ClientID)
	//opts.SetUsername("emqx")
	//opts.SetPassword("public")

    opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
        fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())

		if len(c.DataBuffer) == 0 {
			c.DataBuffer = make(map[string][][]byte)
		}

		if _, exists := c.DataBuffer[msg.Topic()]; exists {
			c.DataBuffer[msg.Topic()] = append(c.DataBuffer[msg.Topic()], msg.Payload())
		} else {
			c.DataBuffer[msg.Topic()] = [][]byte{
				msg.Payload(),
			}
		}
    })

	c.Client = mqtt.NewClient(opts)
	if token := c.Client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	c.Status = 1
	c.Topics = nil

	log.Printf("Connected to MQTT broker: %s\n", c.Broker)
	return nil
}

func (c *Connection) Subscribe(topic string) {
	if slices.Contains(c.Topics, topic) {
		return
	}

	if token := c.Client.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		return
	}
	c.Topics = append(c.Topics, topic)
	fmt.Printf("Subscribed to topic: %s\n", topic)
}

func (c *Connection) Unsubscribe(topic string) {
    if token := c.Client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
        fmt.Println(token.Error())
        return
    }

	for i, v := range c.Topics {
        if v == topic {
            c.Topics = append(c.Topics[:i], c.Topics[i+1:]...)
        }
    }

    fmt.Printf("Unsubscribed from topic: %s\n", topic)
}

func (c *Connection) SendMessage(topic string, message string) {
	token := c.Client.Publish(topic, 0, false, message)
	token.Wait()
	log.Println("Message published:", message)
}

func (c *Connection) Disconnect() {
	c.Client.Disconnect(250)
	c.Status = 0
	log.Printf("Disconnected from MQTT broker: %s\n", c.Broker)
}
