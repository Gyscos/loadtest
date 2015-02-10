package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"sync"
	"time"
)

type Tester struct {
	fileName string
	callRate int
	host     string
}

func NewTester(host string, dataFileName string, callRate int) *Tester {
	return &Tester{
		fileName: dataFileName,
		callRate: callRate,
		host:     host,
	}
}

func (t *Tester) Test() error {
	// Create the request channel
	rc := make(chan string, 20)
	tc := make(chan time.Duration, 20)
	ec := make(chan error, 20)
	times := make([]time.Duration, 0)

	var callGroup sync.WaitGroup
	var handlerGroup sync.WaitGroup

	// Pipe the file into the channel
	handlerGroup.Add(2)
	go t.readFile(rc)
	go handleErrors(ec, &handlerGroup)
	go storeTimes(tc, &times, &handlerGroup)

	// Now read from this channel
	interval := time.Minute / time.Duration(t.callRate)
	for url := range rc {
		callGroup.Add(1)
		go t.runCall(url, tc, ec, &callGroup)
		<-time.After(interval)
	}
	log.Println("All calls sent. Now waiting for them to complete...")
	callGroup.Wait()
	close(ec)
	close(tc)
	handlerGroup.Wait()

	// Now show stats
	showStats(times)

	return nil
}

func showStats(times []time.Duration) {
	avg := 0 * time.Second
	min := 9 * time.Hour
	max := 0 * time.Second
	for _, t := range times {
		avg += t
		if t < min {
			min = t
		}
		if t > max {
			max = t
		}
	}
	avg = avg / time.Duration(len(times))

	variance := 0 * time.Second
	for _, t := range times {
		variance += (t - avg) * (t - avg)
	}
	stdDev := time.Duration(math.Sqrt(float64(variance)))

	log.Println("Average time:", avg)
	log.Println("Min,max:", min, max)
	log.Println("Stddev:", stdDev)
}

func handleErrors(ec <-chan error, wg *sync.WaitGroup) {
	for err := range ec {
		log.Println("Error:", err)
	}
	wg.Done()
}

func storeTimes(tc <-chan time.Duration, times *[]time.Duration, wg *sync.WaitGroup) {
	for t := range tc {
		*times = append(*times, t)
	}
	wg.Done()
}

func (t *Tester) readFile(rc chan<- string) error {
	defer close(rc)

	file, err := os.Open(t.fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		log.Println("Adding", line)
		rc <- line
	}

	return scanner.Err()
}

func (t *Tester) runCall(url string, tc chan<- time.Duration, ec chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()
	resp, err := http.Get(t.host + url)
	if err != nil {
		// Error during call?...
		ec <- err
		return
	}
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		ec <- err
		return
	}

	tc <- time.Since(start)
}
