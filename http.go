package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type ResponsePV struct {
	Values map[int]float32 `json:"values"`
	Time   time.Time       `json:"time"`
}

type ResponseStatus struct {
	Results map[string]float32 `json:"results"`
	Time    time.Time          `json:"time"`
}

func httpStart() {
	http.HandleFunc("/api/v1", httpHandleApi)
	http.HandleFunc("/api/v1/pcs", httpHandleApiPCS)

	http.ListenAndServe(":8080", nil)
}

func httpHandleApiPCS(w http.ResponseWriter, req *http.Request) {
	if lastPCSresponse == nil {
		fmt.Fprintf(w, "no data yet\n")
		return
	}

	q := req.URL.Query()

	fmt.Fprintln(w, time.Now().Format(time.RFC3339))
	fmt.Fprintln(w, lastPCSresponse.ByteCount)

	if q.Get("decode") != "" {
		fmt.Fprintln(w, "decoded")

		for i := 0; i < len(lastPCSresponse.Registers); i += 2 {
			v := int32(lastPCSresponse.Registers[i])<<16 + int32(lastPCSresponse.Registers[i+1])
			fmt.Fprintf(w, "%d\t%d\n", i, v)
		}
	} else {
		fmt.Fprintln(w, "raw")

		for i, v := range lastPCSresponse.Registers {
			fmt.Fprintf(w, "%d\t%d\n", i, v)
		}
	}
}

func httpHandleApi(w http.ResponseWriter, req *http.Request) {
	enc := json.NewEncoder(w)

	results := map[string]float32{}

	for key, result := range lastResults {
		results[key] = result.Value
	}

	resp := ResponseStatus{
		Time:    time.Now(),
		Results: results,
	}

	err := enc.Encode(&resp)
	if err != nil {
		log.Printf("Failed to write reponse: %s\n", err)
	}
}
