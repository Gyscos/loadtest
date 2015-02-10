package main

import (
	"flag"
	"log"
)

func main() {
	var dataFileName string
	var callRate int
	var host string

	flag.StringVar(&host, "h", "http://localhost:8080/api/", "Host base URL")
	flag.StringVar(&dataFileName, "f", "data.txt", "Data file containing list of paths to call")
	flag.IntVar(&callRate, "r", 60, "Call rate: number of calls per minute")
	flag.Parse()

	tester := NewTester(host, dataFileName, callRate)
	err := tester.Test()
	if err != nil {
		log.Println("Error occured:", err)
	}
}
