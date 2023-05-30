// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build !linux
// +build !linux

package main

import "errors"

func monitor(int, chan Message) error {
	return errors.New("not supported")
}
