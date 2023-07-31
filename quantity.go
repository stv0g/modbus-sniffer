// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/binary"
	"fmt"

	"golang.org/x/exp/slog"
)

var ByteOrder = binary.BigEndian

type Quantity struct {
	Register uint16  `json:"register" yaml:"register"`
	Size     int     `json:"size" yaml:"size"`
	Scale    float32 `json:"scale" yaml:"scale"`
	Offset   float32 `json:"offset,omitempty" yaml:"offset,omitempty"`
}

func (q *Quantity) Decode(regs []uint16) (Result, error) {
	if len(regs) != q.Size {
		return Result{}, ErrNotEnoughRegisters
	}

	var fval float32

	switch q.Size {
	case 1:
		fval = float32(int16(regs[0]))

	case 2:
		fval = float32(int32(uint32(regs[0])<<16 + uint32(regs[1])))

	case 4:
		fval = float32(int64(uint64(regs[0])<<48 + uint64(regs[1])<<32 + uint64(regs[2])<<16 + uint64(regs[3])))
	}

	fval += q.Offset
	fval *= q.Scale

	r := Result{
		Quantity: *q,
		Raw:      regs,
		Value:    fval,
	}

	return r, nil
}

type Result struct {
	Quantity Quantity `json:"quantity"`
	Value    float32  `json:"value"`
	Raw      []uint16 `json:"raw"`
}

func (r Result) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("register", fmt.Sprintf("%#x", r.Quantity.Register)),
		slog.String("value", fmt.Sprintf("%.3f", r.Value)),
	)
}
