#!/bin/bash

# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0

#
# start/stop modbus-sniffer daemons.

### BEGIN INIT INFO
# Provides:          modbus-sniffer
# Required-Start:    $network
# Required-Stop:     $network
# Default-Start:     5
# Default-Stop:      0 1 6
# Short-Description: NFC daemon
# Description:       modbus-sniffer is a daemon to publish energy measurements via MQTT
### END INIT INFO

DAEMON="/usr/bin/modbus-sniffer"
PIDFILE="/var/run/modbus-sniffer-*.pid"
DESC="LG ESS MQTT Publisher"

if [ -f /etc/default/modbus-sniffer ] ; then
	. /etc/default/modbus-sniffer
fi

set -e

do_start() {
	start-stop-daemon -S -b -m -p $1 -x ${DAEMON} -- "${@:2}"
}

do_stop() {
	start-stop-daemon -K -p $1
}


do_start_all() {
	do_start ${PIDFILE//\*/pcs} ${ARGS_PCS}
	do_start ${PIDFILE//\*/pm} ${ARGS_PM}
}

do_stop_all() {
	do_stop ${PIDFILE//\*/pcs}
	do_stop ${PIDFILE//\*/pm}
}

case "$1" in
  start)
	echo "Starting ${DESC}"
	do_start_all
	;;

  stop)
	echo "Stopping ${DESC}"
	do_stop_all
	;;

  restart|force-reload)
	echo "Restarting ${DESC}"
	do_stop_all
	sleep 1
	do_start_all
	;;

  *)
	echo "Usage: $0 {start|stop|restart}" >&2
	exit 1
	;;
esac

exit 0
