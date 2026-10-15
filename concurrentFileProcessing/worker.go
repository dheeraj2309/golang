package main

import (
	"strings"
	"sync"
)

type Result struct {
	LineNum    int
	StatusCode string
	Method     string
}

func worker(id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	// The worker constantly pulls jobs off the conveyor belt.
	// This loop automatically pauses if the belt is empty.
	// It automatically breaks and exits ONLY when the jobs channel is closed AND empty.
	for job := range jobs {
		data := strings.Split(job.RawData, " ")
		var method, status string
		if len(data) >= 3 {
			method = data[0]
			status = data[2]
		} else {
			status = "INVALID"
		}

		results <- Result{
			LineNum:    id,
			StatusCode: status,
			Method:     method,
		}
	}
}


