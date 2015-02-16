package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Tester struct {
	fileName   string
	callRate   int
	maxQueries int
	host       string
	w          io.Writer
}

func NewTester(host string, dataFileName string, callRate int, maxQueries int, w io.Writer) *Tester {
	return &Tester{
		fileName:   dataFileName,
		callRate:   callRate,
		maxQueries: maxQueries,
		host:       host,
		w:          w,
	}
}

func (t *Tester) Test(sc <-chan os.Signal) error {
	// sc is the Signal Channel ^^^

	fmt.Fprintln(t.w, "CallRate:", t.callRate)

	// Request channel
	rc := make(chan string, 20)
	// Times channel
	tc := make(chan time.Duration, 20)
	// Error channel
	ec := make(chan error, 20)
	// Abort channel
	ac := make(chan struct{}, 1)

	times := make([]time.Duration, 0)

	// Pipe the file into the channel
	err := t.readFile(rc, ec, sc, ac)
	if err != nil {
		return err
	}

	var callGroup sync.WaitGroup
	var handlerGroup sync.WaitGroup

	handlerGroup.Add(2)
	go t.handleErrors(ec, &handlerGroup)
	go storeTimes(tc, &times, &handlerGroup)

	// Now read from this channel
	interval := time.Minute / time.Duration(t.callRate)
	nCalls := 0
Loop:
	for url := range rc {
		callGroup.Add(1)
		log.Println("Calling", url)
		go t.runCall(url, tc, ec, &callGroup)
		nCalls++
		select {
		case <-ac:
			break Loop
		case <-time.After(interval):
		}
	}
	log.Println("All calls sent. Now waiting for them to complete...")
	callGroup.Wait()
	log.Println("All calls complete. Now closing channels and computing stats.")
	close(ec)
	close(tc)
	handlerGroup.Wait()

	// Now show stats
	nDropped := nCalls - len(times)
	fmt.Fprintf(t.w, "Dropped %v of %v instances (%v %%)\n", nDropped, nCalls, 100*nDropped/nCalls)
	showStats(times, t.w)

	return nil
}

func (t *Tester) handleErrors(ec <-chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	for err := range ec {
		log.Println("Error:", err)
	}
}

func storeTimes(tc <-chan time.Duration, times *[]time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	for t := range tc {
		*times = append(*times, t)
	}
}

func (t *Tester) readFile(rc chan<- string, ec chan<- error, sc <-chan os.Signal, ac chan<- struct{}) error {

	file, err := os.Open(t.fileName)
	if err != nil {
		return err
	}
	go func() {
		defer close(rc)
		defer file.Close()

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			line := scanner.Text()
			select {
			case <-sc:
				log.Println("Stopped reading file.")
				ac <- struct{}{}
				return
			case rc <- line:
				if t.maxQueries > 0 {
					if t.maxQueries == 1 {
						return
					} else {
						t.maxQueries--
					}
				}
			}
		}

		err = scanner.Err()
		if err != nil {
			ec <- err
		}
	}()
	return nil
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
