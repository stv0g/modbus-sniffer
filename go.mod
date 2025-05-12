// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

module github.com/stv0g/modbus-sniffer

go 1.23.0

toolchain go1.24.3

require (
	github.com/eclipse/paho.mqtt.golang v1.5.0
	github.com/howeyc/crc16 v0.0.0-20171223171357-2b2a61e366a6
	golang.org/x/exp v0.0.0-20250408133849-7e4ce0ab07d0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	golang.org/x/net v0.31.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
)
