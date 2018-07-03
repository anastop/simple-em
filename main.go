package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

var elementsPerInterval int64
var cyclesPerElement float64

type HonestMetrics struct {
	CyclesPerElement    float64
	ElementsPerInterval int64
}

type LierTranscoderMetrics struct {
	TranscodingMbps float64
}

type LierStreamerMetrics struct {
	StreamingKbps float64
}

var hmet = &HonestMetrics{}
var tmet = &LierTranscoderMetrics{}
var smet = &LierStreamerMetrics{}

func HonestHandler(w http.ResponseWriter, r *http.Request) {
	hmet.CyclesPerElement = cyclesPerElement
	hmet.ElementsPerInterval = elementsPerInterval

	res, err := json.Marshal(hmet)
	if err != nil {
		log.Fatalf("failed to marshal %v: %v\n", hmet, err)
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(res); err != nil {
		log.Fatalf("failed to write data %+v: %v", res, err)
	}
}

func LierTranscoderHandler(w http.ResponseWriter, r *http.Request) {
	tmet.TranscodingMbps = float64(elementsPerInterval) / (20.0 * 1000000.0)

	res, err := json.Marshal(tmet)
	if err != nil {
		log.Fatalf("failed to marshal %v: %v\n", tmet, err)
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(res); err != nil {
		log.Fatalf("failed to write data %+v: %v", res, err)
	}
}

func LierStreamerHandler(w http.ResponseWriter, r *http.Request) {
	smet.StreamingKbps = float64(elementsPerInterval) / 1000000.0

	res, err := json.Marshal(smet)
	if err != nil {
		log.Fatalf("failed to marshal %v: %v\n", smet, err)
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(res); err != nil {
		log.Fatalf("failed to write data %+v: %v", res, err)
	}
}

func scanner(exp string) {
	var err error
	in := bufio.NewScanner(os.Stdin)
	re := regexp.MustCompile(exp)

	for in.Scan() {
		log.Printf("[%s]\n", in.Text())
		for _, match := range re.FindAllStringSubmatch(in.Text(), -1) {
			elementsPerInterval, err = strconv.ParseInt(match[1], 10, 64)
			if err != nil {
				fmt.Println("conversion error")
			}

			cyclesPerElement, err = strconv.ParseFloat(match[1], 64)
			if err != nil {
				fmt.Println("conversion error")
			}
		}
	}
}

// Run example:
// /home/ubuntu/randacc 35 2>&1 >/dev/null | simple-em
func main() {
	go scanner(`elements_processed:(.+), cycles_per_element:(.+),`)
	//http.HandleFunc("/v1/data", HonestHandler)
	//log.Fatal(http.ListenAndServe(":8000", nil))

	honestMux := http.NewServeMux()
	honestMux.HandleFunc("/v1/data", HonestHandler)
	honestSrv := &http.Server{
		Addr:    ":8000",
		Handler: honestMux,
	}
	go honestSrv.ListenAndServe()

	lierTranscoderMux := http.NewServeMux()
	lierTranscoderMux.HandleFunc("/v1/data", LierTranscoderHandler)
	lierTranscoderSrv := &http.Server{
		Addr:    ":8001",
		Handler: lierTranscoderMux,
	}
	go lierTranscoderSrv.ListenAndServe()

	lierStreamerMux := http.NewServeMux()
	lierStreamerMux.HandleFunc("/v1/data", LierStreamerHandler)
	lierStreamerSrv := &http.Server{
		Addr:    ":8002",
		Handler: lierStreamerMux,
	}
	lierStreamerSrv.ListenAndServe()

}
