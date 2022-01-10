package main

import (
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTClient struct {
	mqtt.Client
}

func mqttConnect(opts *mqtt.ClientOptions) (*MQTTClient, error) {
	client := &MQTTClient{}

	opts.DefaultPublishHandler = client.OnMessage
	opts.OnConnect = client.OnConnect
	opts.OnConnectionLost = client.OnConnectionLost

	client.Client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return client, nil
}

func (c *MQTTClient) OnMessage(_ mqtt.Client, m mqtt.Message) {
	log.Printf("Received MQTT message: %s\n", m.Payload())
}

func (c *MQTTClient) OnConnect(_ mqtt.Client) {
	log.Print("Connected to broker\n")
}

func (c *MQTTClient) OnConnectionLost(_ mqtt.Client, err error) {
	log.Printf("Connection to broker lost: %s\n", err)
}
