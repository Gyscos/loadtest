package main

import (
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
)

func main() {
	var dataFileName string
	var callRate int
	var host string
	var output string

	flag.StringVar(&output, "o", "", "Output file. Leave blank for stdout.")
	flag.StringVar(&host, "h", "http://localhost:8080", "Host base URL")
	flag.StringVar(&dataFileName, "f", "data.txt", "Data file containing list of paths to call")
	flag.IntVar(&callRate, "r", 60, "Call rate: number of calls per minute")
	flag.Parse()

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

	tester := NewTester(host, dataFileName, callRate, w)
	err := tester.Test(c)
	if err != nil {
		log.Println("Error occured:", err)
	}
}
