package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var names = []string{
	"Maia",
	"Reegan",
	"Kelly",
	"Matt",
	"Krisha",
	"Patrick",
	"Tony",
	"Jason",
}

// Envelope for consitent response shape
type envelope struct {
	ElapsedSeconds float64  `json:"elapsed_seconds"`
	Greetings      []string `json:"greetings"`
}

// Create a new envelope and return the pointer
func newEnvelope() *envelope {
	return &envelope{
		Greetings: make([]string, 0),
	}
}

// Example of a request operating on data synchronously
func noWorkerHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	e := newEnvelope()

	/* channels are created for sharing data between processes.
	These are only here because I am lazy and didn't want to write
	another worker */
	jobs := make(chan string, len(names))
	results := make(chan string, len(names))

	// put names into the channel
	for _, name := range names {
		jobs <- name
	}

	// closing jobs channel signals that no more data will be added
	close(jobs)

	// create a worker that will read from jobs and write to resutls
	worker(1, jobs, results)

	// read from the results channel
	for i := 0; i < len(names); i++ {
		e.Greetings = append(e.Greetings, <-results)
	}

	// closing results channel signals that no more data will be read
	close(results)

	e.ElapsedSeconds = time.Since(now).Seconds()
	respJson, err := json.Marshal(e)
	if err != nil {
		w.WriteHeader(http.StatusTeapot)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(respJson)
}

/* worker function that read from the jobs channel and write to the results channel.
Note: the for loop is broken when close(jobs) is called */
func worker(workerId int, jobs <-chan string, results chan<- string) {
	for job := range jobs {
		time.Sleep(1 * time.Second)
		results <- fmt.Sprintf("Hello %s from worker number %d", job, workerId)
	}
}

func workerHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	e := newEnvelope()

	/* channels are created for sharing data between processes.
	These are only here because I am lazy and didn't want to write
	another worker */
	jobs := make(chan string, len(names))
	results := make(chan string, len(names))

	// create 4 background processes (go functions) to process jobs
	for i := 0; i < 5; i++ {
		go worker(i, jobs, results)
	}

	// add jobs to the channel
	for _, name := range names {
		jobs <- name
	}

	// close jobs to signal that no more data will be added
	close(jobs)

	// read from the results channel and add to the greetings
	for i := 0; i < len(names); i++ {
		e.Greetings = append(e.Greetings, <-results)
	}

	// close results channel signaling that no more data will be read
	close(results)

	e.ElapsedSeconds = time.Since(now).Seconds()
	respJson, err := json.Marshal(e)
	if err != nil {
		w.WriteHeader(http.StatusTeapot)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(respJson)
}

func main() {

	http.HandleFunc("/", noWorkerHandler)
	http.HandleFunc("/workers", workerHandler)

	http.ListenAndServe(":3000", nil)
}
