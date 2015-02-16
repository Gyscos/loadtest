package main

import (
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	var dataFileName string
	var callRate int
	var host string
	var output string
	var maxQueries int
	var maxDuration string

	flag.StringVar(&output, "o", "", "Output file. Leave blank for stdout.")
	flag.StringVar(&host, "host", "http://localhost:8080", "Host base URL")
	flag.StringVar(&dataFileName, "f", "data.txt", "Data file containing list of paths to call")
	flag.StringVar(&maxDuration, "maxT", "", "Maximum duration for the test. Leave blank to use maxQ instead.")
	flag.IntVar(&maxQueries, "maxQ", 0, "Maximum number of queries to call. 0 for unlimited.")
	flag.IntVar(&callRate, "r", 60, "Call rate: number of calls per minute")
	flag.Parse()

	if maxDuration != "" && maxDuration != "0" {
		dur, err := time.ParseDuration(maxDuration)
		if err == nil {
			maxQueries = int((time.Duration(callRate) * dur) / time.Minute)
		}
	}

	var w io.Writer
	w = os.Stdout
	if output != "" {
		file, err := os.Create(output)
		if err != nil {
			log.Println("Error occured:", err)
		} else {
			w = file
			defer file.Close()
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	tester := NewTester(host, dataFileName, callRate, maxQueries, w)
	err := tester.Test(c)
	if err != nil {
		log.Println("Error occured:", err)
	}
}
