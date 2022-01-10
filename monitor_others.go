//go:build !linux
// +build !linux

package main

import "errors"

func monitor(int, chan Message) error {
	return errors.New("not supported")
}
