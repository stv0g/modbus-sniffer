// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

type Filter interface {
	Filter(req *ReadHoldingRegistersRequest, resp *ReadHoldingRegistersResponse) bool
}

type PCSFilter struct {
	lastResponse *ReadHoldingRegistersResponse
}

func (p *PCSFilter) Filter(req *ReadHoldingRegistersRequest, resp *ReadHoldingRegistersResponse) bool {
	if p.lastResponse == nil {
		p.lastResponse = resp
		return false
	}

	if len(resp.Registers) < 92 {
		return false
	}

	if req.Address != 0x9c72 {
		return false
	}

	if int32(resp.Registers[50])<<16+int32(resp.Registers[51]) <= 0 {
		return false
	}

	return true
}
