// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build linux
// +build linux

package main

import (
	"errors"
	"fmt"
	"log"
	"runtime"
	"syscall"
	"time"
)

func waitStopped(pid int) error {
	for {
		var sts syscall.WaitStatus

		if _, err := syscall.Wait4(pid, &sts, 0, nil); err != nil {
			return fmt.Errorf("failed to wait: %w", err)
		}

		log.Printf("Process %d stopped by signal %s\n", pid, sts.StopSignal().String())
		if sts.StopSignal() == syscall.SIGSTOP || sts.StopSignal() == syscall.SIGTRAP {
			break
		}
	}

	return nil
}

func monitor(pid int, msgs chan Message) error {
	// https://github.com/golang/go/issues/7699
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	log.Printf("Attaching to %d\n", pid)

	if err := syscall.PtraceAttach(pid); err != nil {
		return fmt.Errorf("failed to attach: %w", err)
	}
	defer syscall.PtraceDetach(pid)

	// Wait for process to be stopped
	if err := waitStopped(pid); err != nil {
		return fmt.Errorf("failed to stop process")
	}

	// Set options
	if err := syscall.PtraceSetOptions(pid, syscall.PTRACE_O_TRACESYSGOOD); err != nil {
		return fmt.Errorf("failed to set ptrace options: %w", err)
	}

	exit := false
	var sig syscall.Signal = 0

	var regs, orig_regs syscall.PtraceRegs

	for {
		// Restart tracee and wait for signal or next syscall
		if err := syscall.PtraceSyscall(pid, int(sig)); err != nil {
			return fmt.Errorf("failed to ptrace syscall: %w", err)
		}

		// Wait for the tracee to stop
		var sts syscall.WaitStatus

		if _, err := syscall.Wait4(pid, &sts, 0, nil); err != nil {
			return fmt.Errorf("failed to wait: %w", err)
		}

		if sts.Stopped() {
			switch sts.StopSignal() {
			case syscall.SIGTRAP | 0x80:
				if err := syscall.PtraceGetRegs(pid, &regs); err != nil {
					return fmt.Errorf("failed to get regs: %w", err)
				}
				if exit {
					handle_syscall(pid, regs, orig_regs, msgs)
				} else {
					orig_regs = regs
				}

				sig = 0

				exit = !exit

			case syscall.SIGSTOP:
				// ignore
			default:
				sig = sts.StopSignal()
			}
		} else {
			return errors.New("wait returned without tracee beeing stopped")
		}
	}
}

func handle_syscall(pid int, regs syscall.PtraceRegs, orig_regs syscall.PtraceRegs, msgs chan Message) {
	var len int
	var dir Direction

	syscall_id, fd, buf, count, ret := decode_syscall_regs(regs, orig_regs)

	switch syscall_id {
	case syscall.SYS_READ:
		len = int(ret)
		dir = DirectionRead

	case syscall.SYS_WRITE:
		len = int(count)
		dir = DirectionWrite

	default:
		return
	}

	if len <= 0 || len > 1<<12 {
		return
	}

	var data []byte = make([]byte, len)
	syscall.PtracePeekData(pid, buf, data)

	// log.Printf("Syscall: id=%d, fd=%d, buf=%x, count=%d, ret=%d, data=%s\n", syscall_id, fd, buf, count, ret, hex.EncodeToString(data))

	msgs <- Message{
		Time:      time.Now(),
		Pid:       pid,
		Fd:        fd,
		Direction: dir,
		Buffer:    data,
	}
}
