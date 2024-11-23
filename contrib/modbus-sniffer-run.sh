#!/bin/bash

# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0

while true; do
	sleep 10
	/usr/bin/modbus-sniffer "$@"
done
