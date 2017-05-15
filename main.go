package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/codahale/hdrhistogram"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"
	"time"
)

type IMSLatency struct {
	Q99, Q95, Q90, Q75, Q50 int64
}

var (
	hist *hdrhistogram.Histogram
	lat  IMSLatency
)

func HTTPHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lat)
}

func record() {
	lat.Q99 = hist.ValueAtQuantile(99.0)
	lat.Q95 = hist.ValueAtQuantile(95.0)
	lat.Q90 = hist.ValueAtQuantile(90.0)
	lat.Q75 = hist.ValueAtQuantile(75.0)
	lat.Q50 = hist.ValueAtQuantile(50.0)
	fmt.Printf("%+v (%d)\n", lat, hist.TotalCount())
}

func recorder(c chan os.Signal, d chan int64, interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	var latency int64
	for {
		select {
		case latency = <-d:
			hist.RecordValue(latency)

		case <-ticker.C:
			record()
			hist.Reset()

		case <-c:
			os.Exit(0)
		}
	}

}

func scanner(exp string, d chan int64) {
	in := bufio.NewScanner(os.Stdin)
	re := regexp.MustCompile(exp)

	for in.Scan() {
		for _, match := range re.FindAllStringSubmatch(in.Text(), -1) {
			//fmt.Printf("%s,", match[1])
			latency, err := strconv.ParseInt(match[1], 10, 64)
			if err != nil {
				fmt.Println("Conversion error")
			}
			d <- latency
		}

	}
}

func main() {
	updateInterval := flag.Int("update-interval", 10,
		"Interval after which the histogram stats should be cleared and populated again")
	flag.Parse()

	sigs := make(chan os.Signal, 1)
	data := make(chan int64)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// max assumed latency: 10 sec, min assumed latency: 1 usec
	hist = hdrhistogram.New(1, 10000000, 2)

	go recorder(sigs, data, *updateInterval)
	go scanner(`Request latency = (\d+) us`, data)

	http.HandleFunc("/v1/data", HTTPHandler)
	log.Fatal(http.ListenAndServe(":9000", nil))
}
