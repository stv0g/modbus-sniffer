// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

func decode_syscall_regs(regs syscall.PtraceRegs) (int, int, uintptr, uintptr, uintptr) {
	var syscall_id int = int(orig_regs.Rax)
	var fd int = int(orig_regs.Rdi)
	var buf uintptr = uintptr(orig_regs.Rsi)
	var count uintptr = uintptr(orig_regs.Rdx)
	var ret uintptr = uintptr(regs.Rax)

	return syscall_id, fd, buf, count, ret
}
