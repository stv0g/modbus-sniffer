package main

import (
	"encoding/binary"
	"fmt"
	"log"

	"github.com/howeyc/crc16"
)

type Decoder struct {
	responseBuffer []byte
	requestBuffer  []byte

	lastRequest *ReadHoldingRegistersRequest
	quantities  map[uint16]Quantity

	Filter func(req *ReadHoldingRegistersRequest, resp *ReadHoldingRegistersResponse) bool
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

	var fc = b[1]
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

	var fc = b[1]
	var cnt = b[2]

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

func NewDecoder(quants map[uint16]Quantity) *Decoder {
	return &Decoder{
		responseBuffer: []byte{},
		requestBuffer:  []byte{},
		lastRequest:    nil,
		quantities:     quants,
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
				log.Printf("Error: %s\n", err)
			}
			return nil
		}

		d.lastRequest = rr

		log.Printf("Req: %#+v\n", rr)

		d.requestBuffer = rem
		d.responseBuffer = []byte{}

	case DirectionRead:
		d.responseBuffer = append(d.responseBuffer, m.Buffer...)

		rr, rem, err := NewReadHoldingRegistersResponse(d.responseBuffer)
		if err != nil {
			if err != ErrNotEnoughData {
				log.Printf("Error: %s\n", err)
			}
			return nil
		}

		log.Printf("Resp: %#+v\n", rr)

		if d.Filter != nil && !d.Filter(d.lastRequest, rr) {
			log.Printf("Skipping filtered response\n")
		}

		for addr, quant := range d.quantities {
			var off int = int(addr) - int(d.lastRequest.Address)
			if off >= 0 && off+quant.Size <= len(rr.Registers) {
				regs := rr.Registers[off : off+quant.Size]

				result, err := quant.Decode(regs)
				if err != nil {
					log.Printf("Error: %s\n", err)
					return nil
				}

				results = append(results, result)
			}
		}

		d.responseBuffer = rem
	}

	return results
}
