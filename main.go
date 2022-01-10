package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	ErrNotEnoughData      = fmt.Errorf("not enough data")
	ErrNotEnoughRegisters = fmt.Errorf("not enough registers")

	pidPCS, pidPM    int
	fromFile, toFile string

	mqttDiscovery bool
	mqttBroker    string
	mqttOpts      *mqtt.ClientOptions = mqtt.NewClientOptions()

	hassioMQTTDiscoveryPrefix = "homeassistant"
	hassioMQTTNodeID          = "lg-ess"

	deviceInfo *Device
)

func openReader(fn string) (*csv.Reader, error) {
	fh, err := os.OpenFile(fn, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	return csv.NewReader(fh), nil
}

func openWriter(fn string) (*csv.Writer, error) {
	fh, err := os.OpenFile(fn, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	return csv.NewWriter(fh), nil
}

func main() {
	var err error

	var reader *csv.Reader
	var writer *csv.Writer

	flag.IntVar(&pidPM, "pid-pm", -1, "PowerMeterMgr pid")
	flag.IntVar(&pidPCS, "pid-pcs", -1, "PCSMgr pid")
	flag.StringVar(&fromFile, "from-file", "", "Read data from file")
	flag.StringVar(&toFile, "to-file", "", "Write data to file")

	flag.StringVar(&mqttOpts.ClientID, "mqtt-client-id", "lg-ess", "MQTT client ID")
	flag.StringVar(&mqttOpts.Username, "mqtt-username", "", "MQTT username")
	flag.StringVar(&mqttOpts.Password, "mqtt-password", "", "MQTT password")
	flag.StringVar(&mqttBroker, "mqtt-broker", "", "MQTT broker url")
	flag.BoolVar(&mqttDiscovery, "mqtt-discovery", true, "Send discovery messages to MQTT")

	flag.StringVar(&hassioMQTTDiscoveryPrefix, "hassio-mqtt-discovery-prefix", "homeassistant", "MQTT Discovery Prefix")
	flag.StringVar(&hassioMQTTNodeID, "hassio-mqtt-node-id", "lg-ess", "MQTT Node ID")

	flag.Parse()

	if mqttOpts.Username == "" {
		log.Fatal("Please provide an MQTT username via the -mqtt-username flag")
	}

	if mqttOpts.Password == "" {
		log.Fatal("Please provide an MQTT password via the -mqtt-password flag")
	}

	if mqttBroker == "" {
		log.Fatal("Please provide an MQTT broker URL via the -mqtt-broker flag")
	}

	mqttOpts.AddBroker(mqttBroker)

	mqttClient, err := mqttConnect(mqttOpts)
	if err != nil {
		log.Fatalf("Failed to connect to MQTT broker: %w", err)
	}

	messages := make(chan Message, 100)

	// Find pids of PCSMgr and PowerMeterMgr
	if pidPCS < 0 {
		pidPCS, err = pidof("PCSMgr")
		if err != nil {
			log.Fatalf("Failed to find pid of PCSMgr: %s", err)
		}

		log.Printf("Detected PID of PCSMgr: %d\n", pidPCS)
	}

	if pidPM < 0 {
		pidPM, err = pidof("PowerMeterMgr")
		if err != nil {
			log.Fatalf("Failed to find pid of PowerMeterMgr: %s", err)
		}

		log.Printf("Detected PID of PowerMeterMgr: %d\n", pidPM)
	}

	pcsQuantities := map[uint16]Quantity{}
	pmQuantities := map[uint16]Quantity{}

	for _, sensor := range Sensors {
		switch sensor.Source {
		case "pcs":
			pcsQuantities[sensor.Quantity.Register] = sensor.Quantity

		case "pm":
			pmQuantities[sensor.Quantity.Register] = sensor.Quantity
		}
	}

	pcs := NewDecoder(pcsQuantities)
	pm := NewDecoder(pmQuantities)

	pcs.Filter = PCSFilter

	if fromFile != "" {
		reader, err = openReader(fromFile)
		if err != nil {
			log.Fatalf("Failed to open file %s: %s", fromFile, err)
		}
	} else {
		reader = nil
	}

	if toFile != "" {
		writer, err = openWriter(toFile)
		if err != nil {
			log.Fatalf("Failed to open file: %s: %s", toFile, err)
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
		for _, pid := range []int{pidPCS, pidPM} {
			go func(pid int) {
				if err := monitor(pid, messages); err != nil {
					log.Fatalf("Failed to ptrace serial communication: %s", err)
				}
			}(pid)
		}
	}

	go httpStart()

	for message := range messages {
		var results []Result
		var source string
		switch message.Pid {
		case pidPCS:
			results = pcs.Decode(message)
			source = "PCS"
		case pidPM:
			results = pm.Decode(message)
			source = "PM"
		}

		for _, result := range results {
			result.Log()

			name := fmt.Sprintf("%s: %s", source, result.Quantity.Name)
			if result.Quantity.Details != "" {
				name = fmt.Sprintf("%s: %s", name, result.Quantity.Details)
			}

			lastResults[name] = result
		}

		if writer != nil {
			message.Write(writer)
		}
	}
}

var lastResults = map[string]Result{}
