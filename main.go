// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"golang.org/x/exp/slog"
)

var (
	ErrNotEnoughData      = fmt.Errorf("not enough data")
	ErrNotEnoughRegisters = fmt.Errorf("not enough registers")
)

var (
	pids []int

	filterMode                                string
	fromFile, toFile, sensorsFile, deviceFile string

	mqttDiscovery bool
	mqttBroker    string
	mqttOpts      *mqtt.ClientOptions = mqtt.NewClientOptions()

	hassioMQTTDiscoveryPrefix string
	hassioMQTTNodeID          string

	httpListenAddr string

	deviceInfo *Device
)

func openReader(fn string) (*csv.Reader, error) {
	fh, err := os.OpenFile(fn, os.O_RDONLY, 0o644)
	if err != nil {
		return nil, err
	}

	return csv.NewReader(fh), nil
}

func openWriter(fn string) (*csv.Writer, error) {
	fh, err := os.OpenFile(fn, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0o644)
	if err != nil {
		return nil, err
	}

	return csv.NewWriter(fh), nil
}

func parseFlags() (err error) {
	flag.StringVar(&fromFile, "from", "", "Read data from file")
	flag.StringVar(&toFile, "to", "", "Write data to file")
	flag.StringVar(&sensorsFile, "sensors", "sensors.yaml", "Sensor definition file")
	flag.StringVar(&deviceFile, "device", "device.yaml", "Device definition file")

	flag.StringVar(&mqttOpts.ClientID, "mqtt-client-id", "modbus-sniffer", "MQTT client ID")
	flag.StringVar(&mqttOpts.Username, "mqtt-username", "", "MQTT username")
	flag.StringVar(&mqttOpts.Password, "mqtt-password", "", "MQTT password")
	flag.StringVar(&mqttBroker, "mqtt-broker", "", "MQTT broker url")
	flag.BoolVar(&mqttDiscovery, "mqtt-discovery", true, "Send discovery messages to MQTT")

	flag.StringVar(&hassioMQTTDiscoveryPrefix, "hassio-mqtt-discovery-prefix", "homeassistant", "MQTT Discovery Prefix")
	flag.StringVar(&hassioMQTTNodeID, "hassio-mqtt-node-id", "modbus-sniffer", "MQTT Node ID")

	flag.StringVar(&httpListenAddr, "http", "", "Listen address for built-in HTTP server")
	flag.StringVar(&filterMode, "filter", "", "Set to 'pcs' to enable PCS filter")

	flag.Parse()

	if mqttBroker != "" {
		mqttOpts.AddBroker(mqttBroker)

		if mqttOpts.Username == "" || mqttOpts.Password == "" {
			return fmt.Errorf("please provide an MQTT username and password via the -mqtt-username, -mqtt-password flags")
		}
	}

	// Get process IDs
	for i := 0; i < flag.NArg(); i++ {
		pidOrProcess := flag.Arg(i)

		var pid int

		if pid, err = strconv.Atoi(pidOrProcess); err != nil {
			if pid, err = pidof(pidOrProcess); err != nil {
				return fmt.Errorf("failed to find pid of process",
					slog.String("process", pidOrProcess),
					slog.Any("error", err))
			}

			slog.Debug("Detected PID of process",
				slog.String("process", pidOrProcess),
				slog.Int("pid", pid))
		}

		pids = append(pids, pid)
	}

	return nil
}

func main() {
	var err error
	var reader *csv.Reader
	var writer *csv.Writer
	var mqttClient *MQTTClient

	if err := parseFlags(); err != nil {
		slog.Error("Failed to parse flags", slog.Any("error", err))
	}

	sensorsList, err := ReadSensors(sensorsFile)
	if err != nil {
		slog.Error("Failed to parse sensor list", slog.Any("error", err))
		return
	}

	if device, err := ReadDevice(deviceFile); err != nil {
		slog.Error("Failed to parse device information", slog.Any("error", err))
		return
	} else if device != nil {
		for i := range sensorsList {
			sensorsList[i].Device = device
		}
	}

	slog.Info("Loaded sensors", slog.Int("count", len(sensorsList)))

	messages := make(chan Message, 100)
	quantities := map[uint16]Quantity{}
	sensors := map[uint16]Sensor{}

	for _, sensor := range sensorsList {
		reg := sensor.Quantity.Register

		quantities[reg] = sensor.Quantity
		sensors[reg] = sensor
	}

	if fromFile != "" {
		reader, err = openReader(fromFile)
		if err != nil {
			slog.Error("Failed to open file", slog.String("file", fromFile), slog.Any("error", err))
			return
		}
	} else {
		reader = nil
	}

	if toFile != "" {
		writer, err = openWriter(toFile)
		if err != nil {
			slog.Error("Failed to open file", slog.String("file", toFile), slog.Any("error", err))
			return
		}
	} else {
		writer = nil
	}

	if reader != nil {
		go func() {
			for {
				message, err := ReadMessage(reader)
				if err != nil {
					log.Fatalf("Failed to read message from file: %s", err)
				}

				messages <- message
			}
		}()
	} else {
		for _, pid := range pids {
			go func(pid int) {
				if err := monitor(pid, messages); err != nil {
					slog.Error("Failed to ptrace serial communication", slog.Any("error", err))
					return
				}
			}(pid)
		}
	}

	if mqttBroker != "" {
		if mqttClient, err = mqttConnect(mqttOpts); err != nil {
			slog.Error("Failed to connect to MQTT broker", slog.Any("error", err))
			return
		}

		mqttClient.WaitUntilConnected()

		if mqttDiscovery {
			for _, sensor := range sensors {
				if err := sensor.SendConfig(mqttClient); err != nil {
					slog.Error("Failed to send MQTT discovery config", slog.Any("error", err))
					return
				}

				slog.Info("Send MQTT discovery config", slog.String("id", sensor.ObjectID))
			}
		}
	}

	if httpListenAddr != "" {
		go httpStart(httpListenAddr)
	}

	var filter Filter
	switch filterMode {
	case "pcs":
		filter = &PCSFilter{}
	}

	dec := NewDecoder(filter, quantities)

	for message := range messages {
		results := dec.Decode(message)

		for _, result := range results {
			reg := result.Quantity.Register

			sensor := sensors[reg]

			slog.Info("New value",
				slog.Int("pid", message.Pid),
				slog.Int("fd", message.Fd),
				slog.Any("result", result), slog.Any("sensor", sensor))

			name := fmt.Sprintf("%#x", reg)
			lastResponseResult[name] = ResponseStatusResult{
				Sensor: sensor,
				Value:  result.Value,
			}

			if mqttClient != nil {
				sensor.SendState(mqttClient, result.Value)
			}
		}

		if writer != nil {
			message.Write(writer)
		}
	}
}
