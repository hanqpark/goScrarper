package main

import (
	"fmt"
	"sync"

	"github.com/hanqpark/jobScraper/scrape"
)

func main() {
	var wg sync.WaitGroup

	totalPages := scrape.GetPages()
	
	ch := make(chan []scrape.ExtractedJob)
	for i:=0; i<totalPages; i++ {
		go scrape.GetPage(i+1, ch)
	}

	w := scrape.CreateFile()
	defer w.Flush()

	for i:=0; i<totalPages; i++ {
		wg.Add(1)
		go scrape.WriteJobs(<-ch, w, &wg)
		// jobs = append(jobs, extractedJobs...)  // 2개의 배열을 합치려면 ... 붙이기
	}

	wg.Wait()

	fmt.Println("Done")
}

