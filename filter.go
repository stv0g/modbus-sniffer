package main

var lastPCSresponse *ReadHoldingRegistersResponse = nil

func PCSFilter(req *ReadHoldingRegistersRequest, resp *ReadHoldingRegistersResponse) bool {
	lastPCSresponse = resp

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
