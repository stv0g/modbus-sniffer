// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v3"
)

const (
	ComponentSensor = "sensor"

	// https://www.home-assistant.io/docs/configuration/customizing-devices/#device-class
	DeviceClassBattery     = "battery"      // Percentage of battery that is left.
	DeviceClassCurrent     = "current"      // Current in A.
	DeviceClassEnergy      = "energy"       // Energy in Wh, kWh or MWh.
	DeviceClassPower       = "power"        // Power in W or kW.
	DeviceClassPowerFactor = "power_factor" // Power factor in %.
	DeviceClassTemperature = "temperature"  // Temperature in °C or °F.
	DeviceClassVoltage     = "voltage"      // Voltage in V.
	DeviceClassFrequency   = "frequency"    // Frequency in Hz, kHz, MHz or GHz.

	// https://developers.home-assistant.io/docs/core/entity/sensor#available-state-classes
	StateClassMeasurement     = "measurement"
	StateClassTotal           = "total"
	StateClassTotalIncreasing = "total_increasing"
)

type Device struct {
	Name         string `json:"name,omitempty" yaml:"name,omitempty"`
	Model        string `json:"model,omitempty" yaml:"model,omitempty"`
	Manufacturer string `json:"manufacturer,omitempty" yaml:"manufacturer,omitempty"`

	ConfigurationURL string `json:"configuration_url,omitempty" yaml:"configuration_url,omitempty"`

	SoftwareVersion string `json:"sw_version,omitempty" yaml:"sw_version,omitempty"`

	// Identifiers  []string `json:"identifiers`
	Connections [][]string `json:"connections,omitempty" yaml:"connections,omitempty"`
}

type Sensor struct {
	Quantity Quantity `json:"modbus" yaml:"modbus"`
	Device   *Device  `json:"device,omitempty" yaml:"device,omitempty"`

	ObjectID          string `json:"object_id,omitempty" yaml:"object_id,omitempty"`
	UniqueID          string `json:"unique_id,omitempty" yaml:"unique_id,omitempty"`
	Name              string `json:"name,omitempty" yaml:"name,omitempty"`
	DeviceClass       string `json:"device_class,omitempty" yaml:"device_class,omitempty"`
	StateClass        string `json:"state_class,omitempty" yaml:"state_class,omitempty"`
	StateTopic        string `json:"state_topic,omitempty" yaml:"state_topic,omitempty"`
	UnitOfMeasurement string `json:"unit_of_measurement,omitempty" yaml:"unit_of_measurement,omitempty"`
	Icon              string `json:"icon,omitempty" yaml:"icon,omitempty"`
	Component         string `json:"component" yaml:"component"`
}

func ReadSensors(fn string) ([]Sensor, error) {
	var sensors []Sensor

	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(&sensors); err != nil {
		return nil, err
	}

	return sensors, nil
}

func ReadDevice(fn string) (*Device, error) {
	device := &Device{}

	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(device); err != nil {
		return nil, err
	}

	return device, nil
}

func (s *Sensor) Topic(sub string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s", hassioMQTTDiscoveryPrefix, s.Component, hassioMQTTNodeID, s.ObjectID, sub)
}

func (s *Sensor) SendConfig(c mqtt.Client) error {
	s.StateTopic = s.Topic("state")

	t := *s
	if t.UniqueID == "" {
		t.UniqueID = t.ObjectID
	}

	payload, err := json.Marshal(&t)
	if err != nil {
		return err
	}

	c.Publish(s.Topic("config"), 2, false, payload)

	return nil
}

func (s *Sensor) SendState(c mqtt.Client, value float32) {
	payload := fmt.Sprintf("%.2f", value)

	c.Publish(s.Topic("state"), 2, false, payload)
}

func (s Sensor) LogValue() slog.Value {
	as := []slog.Attr{}

	if id := s.UniqueID; id != "" {
		as = append(as, slog.String("unique_id", id))
	}

	if id := s.ObjectID; id != "" {
		as = append(as, slog.String("unique_id", id))
	}

	if name := s.Name; name != "" {
		as = append(as, slog.String("name", name))
	}

	if unit := s.UnitOfMeasurement; unit != "" {
		as = append(as, slog.String("unit", unit))
	}

	return slog.GroupValue(as...)
}
