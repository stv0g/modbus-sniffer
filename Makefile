# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
#
# SPDX-License-Identifier: Apache-2.0

HOST=192.168.178.2
USER=root
EXEC=modbus-sniffer

SSH=ssh $(USER)@$(HOST)

export GOOS=linux
export GOARCH=arm

# For loading $(ARGS) used in run target
include contrib/modbus-sniffer

all: build

build:
	go build -o $(EXEC) .

install: build
	$(SSH) killall $(EXEC) || true
	scp $(EXEC) $(USER)@$(HOST):/usr/bin/modbus-sniffer
	$(SSH) mkdir -p /etc/modbus-sniffer
	scp etc/sensors.yaml $(USER)@$(HOST):/etc/modbus-sniffer/sensors.yaml
	scp etc/device.yaml $(USER)@$(HOST):/etc/modbus-sniffer/device.yaml
	scp contrib/modbus-sniffer.sh $(USER)@$(HOST):/etc/init.d/
	scp contrib/modbus-sniffer-run.sh $(USER)@$(HOST):/usr/bin/modbus-sniffer-run.sh
	scp contrib/modbus-sniffer $(USER)@$(HOST):/etc/default/modbus-sniffer
	$(SSH) ln -fs /etc/init.d/modbus-sniffer.sh /etc/rc5.d/S80modbus-sniffer.sh

uninstall:
	$(SSH) killall $(EXEC)
	$(SSH) rm -f \
		/etc/modbus-sniffer \
		/etc/default/modbus-sniffer \
		/usr/bin/modbus-sniffer \
		/usr/bin/modbus-sniffer-run.sh \
		/etc/init.d/modbus-sniffer.sh \
		/etc/rc5.d/S80modbus-sniffer.sh \
		/var/run/modbus-sniffer*.pid

run: install run-directly

run-directly:
	$(SSH) killall $(EXEC) || true
	$(SSH) /usr/bin/$(EXEC) $(ARGS) -filter=pcs PCSMgr

.PHONY: all build run run-directly install uninstall
