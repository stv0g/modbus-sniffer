package main

import "syscall"

func decode_syscall_regs(regs, orig_regs syscall.PtraceRegs) (int, int, uintptr, uintptr, uintptr) {
	var syscall_id int = int(orig_regs.Uregs[7])
	var fd int = int(orig_regs.Uregs[0])
	var buf uintptr = uintptr(orig_regs.Uregs[1])
	var count uintptr = uintptr(orig_regs.Uregs[2])

	var ret uintptr = uintptr(regs.Uregs[0])

	return syscall_id, fd, buf, count, ret
}
