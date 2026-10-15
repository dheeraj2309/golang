package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)


func main(){
	start := time.Now()
	// we make the channels for the coordination of producer,worker and result aggregator
	jobs := make(chan Job, 1000) // string channel with capacity of 1000 
	results := make(chan Result,1000)

	var wg sync.WaitGroup

	numWorkers := runtime.NumCPU()
	fmt.Printf("Deploying %d workers...\n", numWorkers)

	for i := range numWorkers{
		wg.Add(1)
		go worker(i,jobs,results,&wg)
	}

	// The Invisible Monitor (Crucial for Shutdown)
	// This Goroutine sits in the background doing nothing but waiting for the counter to hit 0.
	go func() {
		wg.Wait()      // Blocks until all workers call wg.Done()
		close(results) // Safely closes the results belt ONLY when all workers are gone
	}()
	// Start the Producer (Data Intake)
	// We run this in a background Goroutine so it doesn't block the main thread.
	// It will stream the file, and call close(jobs) when it hits EOF.
	go func(){
		err := streamInput("server.log",jobs)
		if err != nil{
			fmt.Println("Producer Error:", err)
		}
	}()

	// Start the Collector (Fan-In)
	// This runs on the MAIN thread. It blocks the program from exiting until 
	// the results channel is fully closed by the Invisible Monitor.
	finalReport := makeReport(results)

	// 6. Print the Results
	fmt.Println("--- Pipeline Finished ---")
	fmt.Printf("Total Requests: %d\n", finalReport.TotalRequests)
	fmt.Printf("Status Counts:  %v\n", finalReport.StatusCount)
	fmt.Printf("Execution Time: %v\n", time.Since(start))
}