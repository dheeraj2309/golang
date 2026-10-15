package main

type Report struct {
	TotalRequests int
	StatusCount   map[string]int
}

func makeReport(results <-chan Result) Report {
	report := Report{
		TotalRequests: 0,
		StatusCount:   make(map[string]int),
	}
	for res := range results {
		report.TotalRequests++
		report.StatusCount[res.StatusCode]++
	}
	return report
}
