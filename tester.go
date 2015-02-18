package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/Gyscos/urlspammer"
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

func (t *Tester) addHost(output chan<- string, input <-chan string) {
	for path := range input {
		output <- t.host + path
	}
	close(output)
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
	nCalls := 0
	err := t.readFile(rc, ec, sc, ac, &nCalls)
	if err != nil {
		return err
	}
	// Url Channel
	uc := make(chan string, 5)
	go t.addHost(uc, rc)

	var handlerGroup sync.WaitGroup

	handlerGroup.Add(2)
	// Log errors from ec
	go t.handleErrors(ec, &handlerGroup)
	// Store input from tc into times
	go storeTimes(tc, &times, &handlerGroup)

	urlspammer.SpamByRate(t.callRate, urlspammer.WrapUrls(uc), func(q urlspammer.Query, body []byte, d time.Duration) {
		// Called on each successful query
		log.Printf("[%v] %v\n", d, q.Url)
		tc <- d
	})
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

// Read a file and push urls to the channel.
// When a signal is caught, stop, and send somehting along ac, the AbortChannel
func (t *Tester) readFile(rc chan<- string, ec chan<- error, sc <-chan os.Signal, ac chan<- struct{}, nCalls *int) error {

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
				(*nCalls)++
				if t.maxQueries == 0 {
					break
				}
				if t.maxQueries == 1 {
					return
				}
				t.maxQueries--
			}
		}

		err = scanner.Err()
		if err != nil {
			ec <- err
		}
	}()
	return nil
}
