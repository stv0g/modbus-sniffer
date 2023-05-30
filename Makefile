# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
#
# SPDX-License-Identifier: Apache-2.0

HOST=LG-Sun
USER=root
EXEC=lg-ess-mqtt

SSH=ssh $(USER)@$(HOST)

export GOOS=linux
export GOARCH=arm

all: build

build:
	go build -o $(EXEC) .

install: build
	$(SSH) killall $(EXEC)
	scp $(EXEC) $(USER)@$(HOST):/home/root/
	scp contrib/lg-ess-mqtt.sh $(USER)@$(HOST):/etc/init.d/
	scp contrib/lg-ess-mqtt $(USER)@$(HOST):/etc/default/lg-ess-mqtt
	$(SSH) ln -fs /etc/init.d/lg-ess-mqtt.sh /etc/rc5.d/S80lg-ess-mqtt.sh

uninstall:
	$(SSH) killall $(EXEC)
	$(SSH) rm -f /etc/default/lg-ess-mqtt
	$(SSH) rm -f /home/root/lg-ess-mqtt
	$(SSH) rm -f /etc/init.d/lg-ess-mqtt.sh
	$(SSH) rm -f /etc/rc5.d/S80lg-ess-mqtt.sh

run: install
	$(SSH) /home/root/$(EXEC)

.PHONY: all build run install

