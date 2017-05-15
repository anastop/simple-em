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

type Metrics struct {
	CyclesPerElement float64
}

var m = &Metrics{}

func HTTPHandler(w http.ResponseWriter, r *http.Request) {
	res, err := json.Marshal(m)
	if err != nil {
		log.Fatalf("failed to marshal %v: %v\n", m, err)
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
			m.CyclesPerElement, err = strconv.ParseFloat(match[1], 64)
			if err != nil {
				fmt.Println("Conversion error")
			}
		}
	}
}

func main() {
	go scanner(`cycles_per_element:(.+),`)
	http.HandleFunc("/v1/data", HTTPHandler)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
