package main

import (
	"bufio"
	"os"
)

type Job struct {
	LineNum int
	RawData string
}

func streamInput(filePath string, jobs chan<- Job) error { // jobs chan<- Job,It tells the compiler, "Inside this function, this channel is send-only."
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 1

	for scanner.Scan() {
		jobs <- Job{
			LineNum: lineNum,
			RawData: scanner.Text(),
		}
		lineNum++
	}
	close(jobs)
	return scanner.Err()
}
