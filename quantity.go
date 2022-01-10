package main

import "log"

type Decode func(regs []uint16) (Result, error)

type Quantity struct {
	Register uint16 `json:"register"`
	Size         int     `json:"size"`
	Scale        float32 `json:"scale"`
	Offset       float32 `json:"offset"`
	CustomDecode Decode  `json:"-"`
}

func (q *Quantity) Decode(regs []uint16) (Result, error) {
	if q.CustomDecode != nil {
		return q.CustomDecode(regs)
	}

	if len(regs) != q.Size {
		return Result{}, ErrNotEnoughRegisters
	}

	var val int64 = 0
	for i := 0; i < q.Size; i++ {
		val += int64(regs[q.Size-i-1]) << (i * 16)
	}

	r := Result{
		Quantity: *q,
		Value:    (float32(val) + q.Offset) * q.Scale,
	}

	return r, nil
}

type Result struct {
	Quantity Quantity `json:"quantity"`
	Value    float32  `json:"value"`
}

func (r Result) Log() {
	// log.Printf("Quant: %#+v\n", d.lastRequest)
	log.Printf("%s %s [%s]: %f\n", r.Quantity.Name, r.Quantity.Details, r.Quantity.Unit, r.Value)
}
