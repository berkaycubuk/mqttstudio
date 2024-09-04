package main

import (
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Connection struct {
	ProjectID int
	Status int // 0: offline, 1: online
	Broker string
	Client mqtt.Client
}

func (c *Connection) Connect() error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(c.Broker)
	opts.SetClientID("mqtt_studio_dev_client_1")
	opts.SetUsername("emqx")
	opts.SetPassword("public")

	c.Client = mqtt.NewClient(opts)
	if token := c.Client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	c.Status = 1

	log.Printf("Connected to MQTT broker: %s\n", c.Broker)
	return nil
}

func (c *Connection) Disconnect() {
	c.Client.Disconnect(250)
	c.Status = 0
	log.Printf("Disconnected from MQTT broker: %s\n", c.Broker)
}
