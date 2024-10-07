// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

module github.com/stv0g/modbus-sniffer

go 1.22.0

toolchain go1.23.1

require (
	github.com/eclipse/paho.mqtt.golang v1.5.0
	github.com/howeyc/crc16 v0.0.0-20171223171357-2b2a61e366a6
	golang.org/x/exp v0.0.0-20241004190924-225e2abe05e6
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	golang.org/x/net v0.27.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
)
