<!--
SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
SPDX-License-Identifier: Apache-2.0
-->

# modbus-sniffer

`modbus-sniffer` is little Go command line utility which sniffs Modbus communication (read holding registers commands) via Linux's `ptrace` syscall.
It does so by attaching itself to a specified process and intercepting all `read()` & `write()` system calls which are used to communicate to a Modbus device attached to a serial port or TCP network.

This project is used as a firmware extension for the 1st generation LG ESS PV/Battery systems to publish the internal system state periodically via MQTT to Homeassistant.

## Requirements

- 1st generation LG ESS
- [Make](https://www.gnu.org/software/make/)
- [Go](https://go.dev/)
- [OpenSSH SCP](https://www.openssh.com/)

## Tested LG ESS products

- [ED05K000E00](https://www.lg.com/de/business/solar/downloadbereich/datenblaetter/ESS/LG02.3692_ESS_DataSheet_DE_0507_RZ.pdf)

## Configuration

Adjust the flags in `contrib/modbus-sniffer` before running `make install`.

## Usage

```shell
make install
```

## Credits

- Steffen Vogel ([@stv0g](https://github.com/stv0g))
