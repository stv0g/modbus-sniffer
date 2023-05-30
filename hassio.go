// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	DeviceClassBattery     = "battery"      // Percentage of battery that is left.
	DeviceClassCurrent     = "current"      // Current in A.
	DeviceClassEnergy      = "energy"       // Energy in Wh, kWh or MWh.
	DeviceClassPower       = "power"        // Power in W or kW.
	DeviceClassPowerFactor = "power_factor" // Power factor in %.
	DeviceClassTemperature = "temperature"  // Temperature in °C or °F.
	DeviceClassVoltage     = "voltage"      // Voltage in V.
	DeviceClassFrequency   = "frequency"    // Frequency in Hz, kHz, MHz or GHz.

	StateClassMeasurement     = "measurement"
	StateClassTotal           = "total"
	StateClassTotalIncreasing = "total_increasing"
)

type Device struct {
	Name         string `json:"name"`
	Model        string `json:"model"`
	Manufacturer string `json:"manufacturer"`

	SoftwareVersion string `json:"sw_version"`

	// Identifiers  []string `json:"identifiers`
	Connections [][]string `json:"connections"`
}

type Sensor struct {
	ObjectID          string `json:"object_id"`
	Name              string `json:"name"`
	FriendlyName      string `json:"friendly_name,omitempty"`
	DeviceClass       string `json:"device_class,omitempty"`
	StateClass        string `json:"state_class,omitempty"`
	StateTopic        string `json:"state_topic,omitempty"`
	UnitOfMeasurement string `json:"unit_of_measurement,omitempty"`
	Icon              string `json:"icon,omitempty"`

	Device *Device `json:"device,omitempty"`

	Source string `json:"-"`

	Component string `json:"-"`
	Quantity  `json:"-"`
}

func GetDeviceInfo() (*Device, error) {
	iface, err := net.InterfaceByName("eth0")
	if err != nil {
		return nil, err
	}

	conns := [][]string{
		{"mac", iface.HardwareAddr.String()},
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		conns = append(conns, []string{"ip", addr.String()})
	}

	version, err := ioutil.ReadFile("/etc/version")
	if err != nil {
		return nil, err
	}

	return &Device{
		Name:            "LG ESS Gen1",
		Manufacturer:    "LG Electronics Inc.",
		Model:           "ED05K000E00",
		Connections:     conns,
		SoftwareVersion: string(version),
	}, nil
}

func (s *Sensor) Topic() string {
	return fmt.Sprintf("%s/%s/%s/%s", hassioMQTTDiscoveryPrefix, s.Component, hassioMQTTNodeID, s.ObjectID)
}

func (s *Sensor) Discover(c mqtt.Client) error {
	s.StateTopic = s.Topic() + "/state"

	payload, err := json.Marshal(s)
	if err != nil {
		return err
	}

	c.Publish(s.Topic()+"/config", 2, false, payload)

	return nil
}

func (s *Sensor) Update(c mqtt.Client, value float32) {
	payload := fmt.Sprintf("%f", value)

	c.Publish(s.Topic()+"/state", 2, false, payload)
}
