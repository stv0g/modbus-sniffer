// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/tls"
	"net/url"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"golang.org/x/exp/slog"
)

type MQTTClient struct {
	mqtt.Client

	connected sync.WaitGroup
}

func mqttConnect(opts *mqtt.ClientOptions) (*MQTTClient, error) {
	client := &MQTTClient{}

	opts.OnConnectAttempt = func(broker *url.URL, tlsCfg *tls.Config) *tls.Config {
		slog.Info("Attempt connection to broker", slog.Any("broker", broker))

		client.connected.Add(1)

		return tlsCfg
	}

	opts.OnConnect = func(_ mqtt.Client) {
		slog.Info("Connected to broker")
		client.connected.Done()
	}

	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		slog.Info("Connection to broker lost", slog.Any("error", err))
	}

	client.Client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return client, nil
}

func (c *MQTTClient) WaitUntilConnected() {
	c.connected.Wait()
}
