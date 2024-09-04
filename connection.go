package main

import (
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Connection struct {
	ProjectID int
	Status int // 0: offline, 1: online
	Broker string
	Client mqtt.Client
	Topics []string
	DataBuffer map[string][][]byte
}

func (c *Connection) Connect() error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(c.Broker)
	opts.SetClientID("mqtt_studio_dev_client_1")
	opts.SetUsername("emqx")
	opts.SetPassword("public")

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

func (c *Connection) Disconnect() {
	c.Client.Disconnect(250)
	c.Status = 0
	log.Printf("Disconnected from MQTT broker: %s\n", c.Broker)
}
