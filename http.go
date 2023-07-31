// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/exp/slog"
)

var (
	lastReadHoldingRegistersResponse *ReadHoldingRegistersResponse
	lastResponseResult               = map[string]ResponseStatusResult{}
)

func httpStart(addr string) {
	http.HandleFunc("/api/v1/status", httpHandleApiStatus)
	http.HandleFunc("/api/v1/raw", httpHandleApiRaw)

	http.ListenAndServe(addr, nil)
}

type ResponseStatus struct {
	Results map[string]ResponseStatusResult `json:"results"`
	Time    time.Time                       `json:"time"`
}

type ResponseStatusResult struct {
	Sensor
	Value float32 `json:"value"`
}

func httpHandleApiStatus(w http.ResponseWriter, req *http.Request) {
	resp := ResponseStatus{
		Time:    time.Now(),
		Results: lastResponseResult,
	}

	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		slog.Error("Failed to write response", slog.Any("error", err))
	}
}

type ResponseRaw struct {
	Time         string   `json:"time"`
	Unit         byte     `json:"unit"`
	FunctionCode byte     `json:"function_code"`
	ByteCount    byte     `json:"count"`
	Registers    []uint16 `json:"registers"`
	Checksum     uint16   `json:"checksum"`
}

func httpHandleApiRaw(w http.ResponseWriter, req *http.Request) {
	if lastReadHoldingRegistersResponse == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("no data yet\n"))
		return
	}

	resp := ResponseRaw{
		Time:         time.Now().Format(time.RFC3339),
		Unit:         lastReadHoldingRegistersResponse.ByteCount,
		FunctionCode: lastReadHoldingRegistersResponse.FunctionCode,
		ByteCount:    lastReadHoldingRegistersResponse.ByteCount,
		Registers:    lastReadHoldingRegistersResponse.Registers,
		Checksum:     lastReadHoldingRegistersResponse.Checksum,
	}

	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		slog.Error("Failed to write response", slog.Any("error", err))
	}
}
