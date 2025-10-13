// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

module github.com/stv0g/modbus-sniffer

go 1.24.0

toolchain go1.25.2

require (
	github.com/eclipse/paho.mqtt.golang v1.5.1
	github.com/howeyc/crc16 v0.0.0-20171223171357-2b2a61e366a6
	golang.org/x/exp v0.0.0-20251002181428-27f1f14c8bb9
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
)
