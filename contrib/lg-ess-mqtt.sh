#!/bin/sh

# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0

#
# start/stop lg-ess-mqtt daemon.

### BEGIN INIT INFO
# Provides:          lg-ess-mqtt
# Required-Start:    $network
# Required-Stop:     $network
# Default-Start:     5
# Default-Stop:      0 1 6
# Short-Description: NFC daemon
# Description:       lg-ess-mqtt is a daemon to publish energy measurements via MQTT
### END INIT INFO

DAEMON=/home/root/lg-ess-mqtt
PIDFILE=/var/run/lg-ess-mqtt.pid
DESC="LG ESS MQTT Publisher"

if [ -f /etc/default/lg-ess-mqtt ] ; then
	. /etc/default/lg-ess-mqtt
fi

set -e

do_start() {
	start-stop-daemon -S -b -m -p $PIDFILE -x $DAEMON -- $ARGS
}

do_stop() {
	start-stop-daemon -K -p $PIDFILE
}

case "$1" in
  start)
	echo "Starting $DESC"
	do_start
	;;
  stop)
	echo "Stopping $DESC"
	do_stop
	;;
  restart|force-reload)
	echo "Restarting $DESC"
	do_stop
	sleep 1
	do_start
	;;
  *)
	echo "Usage: $0 {start|stop|restart|force-reload}" >&2
	exit 1
	;;
esac

exit 0
