// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/binary"
	"fmt"

	"github.com/howeyc/crc16"
	"golang.org/x/exp/slog"
)

type FilterFunc func(req *ReadHoldingRegistersRequest, resp *ReadHoldingRegistersResponse) bool

type Decoder struct {
	responseBuffer []byte
	requestBuffer  []byte

	lastRequest *ReadHoldingRegistersRequest
	quantities  map[uint16]Quantity
	filter      Filter
}

type ReadHoldingRegistersRequest struct {
	Unit          byte
	FunctionCode  byte
	Address       uint16
	RegisterCount uint16
	Checksum      uint16
}

type ReadHoldingRegistersResponse struct {
	Unit         byte
	FunctionCode byte
	ByteCount    byte
	Registers    []uint16
	Checksum     uint16
}

func NewReadHoldingRegistersRequest(b []byte) (*ReadHoldingRegistersRequest, []byte, error) {
	if len(b) < 8 {
		return nil, nil, ErrNotEnoughData
	}

	fc := b[1]
	if fc != 3 {
		return nil, nil, fmt.Errorf("invalid function code: %d", fc)
	}

	r := &ReadHoldingRegistersRequest{
		Unit:          b[0],
		FunctionCode:  fc,
		Address:       binary.BigEndian.Uint16(b[2:4]),
		RegisterCount: binary.BigEndian.Uint16(b[4:6]),
		Checksum:      binary.LittleEndian.Uint16(b[6:8]),
	}

	if r.Checksum != ^crc16.ChecksumIBM(b[0:6]) {
		return nil, nil, fmt.Errorf("invalid checksum")
	}

	return r, []byte{}, nil
}

func NewReadHoldingRegistersResponse(b []byte) (*ReadHoldingRegistersResponse, []byte, error) {
	len := len(b)
	if len <= 3 {
		return nil, nil, ErrNotEnoughData
	}

	fc := b[1]
	cnt := b[2]

	if fc != 3 {
		return nil, nil, fmt.Errorf("invalid function code: %d", fc)
	}

	if len < int(cnt+5) {
		return nil, nil, ErrNotEnoughData
	}

	r := &ReadHoldingRegistersResponse{
		Unit:         b[0],
		FunctionCode: fc,
		ByteCount:    cnt,
		Registers:    []uint16{},
		Checksum:     binary.LittleEndian.Uint16(b[3+cnt : 5+cnt]),
	}

	var i byte
	for i = 0; i < cnt/2; i++ {
		r.Registers = append(r.Registers, binary.BigEndian.Uint16(b[(i*2)+3:(i*2)+5]))
	}

	if r.Checksum != ^crc16.ChecksumIBM(b[0:3+cnt]) {
		return nil, nil, fmt.Errorf("invalid checksum")
	}

	return r, b[5+cnt:], nil
}

func NewDecoder(filter Filter, quants map[uint16]Quantity) *Decoder {
	return &Decoder{
		quantities: quants,
		filter:     filter,
	}
}

func (d *Decoder) Decode(m Message) []Result {
	results := []Result{}

	switch m.Direction {

	// This is a normal Modbus read holding registers request
	case DirectionWrite:
		rr, rem, err := NewReadHoldingRegistersRequest(m.Buffer)
		if err != nil {
			if err != ErrNotEnoughData {
				slog.Error("Failed to parse read holding register request", slog.Any("error", err))
			}
			return nil
		}

		d.lastRequest = rr

		slog.Debug("ReadHoldingRegistersRequest", slog.Any("addr", rr.Address), slog.Any("count", rr.RegisterCount), slog.Any("unit", rr.Unit))

		d.requestBuffer = rem
		d.responseBuffer = []byte{}

	case DirectionRead:
		d.responseBuffer = append(d.responseBuffer, m.Buffer...)

		if d.lastRequest == nil {
			slog.Error("No request yet")
			return nil
		}

		rr, rem, err := NewReadHoldingRegistersResponse(d.responseBuffer)
		if err != nil {
			if err != ErrNotEnoughData {
				slog.Error("Failed to parse holding registers response", slog.Any("error", err))
			}
			return nil
		}

		regs := []string{}
		for _, r := range rr.Registers {
			regs = append(regs, fmt.Sprint(r))
		}

		slog.Debug("ReadHoldingRegistersResponse", slog.Any("count", rr.ByteCount), slog.Any("unit", rr.Unit), slog.Any("register", regs))

		if d.filter != nil && !d.filter.Filter(d.lastRequest, rr) {
			slog.Debug("Skipping filtered response")
			return nil
		}

		for addr, quant := range d.quantities {
			var off int = int(addr) - int(d.lastRequest.Address)
			if off >= 0 && off+quant.Size <= len(rr.Registers) {
				regs := rr.Registers[off : off+quant.Size]

				result, err := quant.Decode(regs)
				if err != nil {
					slog.Error("Failed to decode quantity", slog.Any("error", err))
					return nil
				}

				results = append(results, result)
			}
		}

		d.responseBuffer = rem
	}

	return results
}
