package scraper

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

// Scrape Saramin by a term
func Scrape(term string) {
	baseURL = fmt.Sprintf("%s&searchword=%s", baseURL, term)
	var wg sync.WaitGroup

	totalPages := getPages()
	
	ch := make(chan []extractedJob)
	for i:=0; i<totalPages; i++ {
		go getPage(i+1, ch)
	}

	w := createFile()
	defer w.Flush()

	for i:=0; i<totalPages; i++ {
		wg.Add(1)
		go writeJobs(<-ch, w, &wg)
		// jobs = append(jobs, extractedJobs...)  // 2개의 배열을 합치려면 ... 붙이기
	}

	wg.Wait()

	fmt.Println("Done")
}

func getPage(page int, chMain chan <- []extractedJob) {
	pageURL := baseURL + "&recruitPage=" + strconv.Itoa(page)
	fmt.Println(pageURL)

	res, err := http.Get(pageURL)
	checkErr(err)
	checkStatusCode(res)
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	ch := make(chan extractedJob)
	searchCards := doc.Find(".item_recruit")
	searchCards.Each(func(i int, card *goquery.Selection){
		go extractJob(card, ch)
	})

	var jobs []extractedJob
	for i:=0; i<searchCards.Length(); i++ {
		job := <- ch
		jobs = append(jobs, job)
	}
	chMain <- jobs
}

func getPages() (pages int) {
	res, err := http.Get(baseURL)
	checkErr(err)
	checkStatusCode(res)
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection){
		pages = s.Find("a").Length()
	})

	return pages
}

func writeJobs(jobs []extractedJob, w *csv.Writer, wg *sync.WaitGroup) {
	defer wg.Done()
	
	chJob := make(chan []string)
	// go routine으로 변형 가능
	for _, job := range jobs {
		go writeJob(chJob, job)
	}

	for i:=0; i<len(jobs); i++ {
		jErr := w.Write(<-chJob)
		checkErr(jErr)
	}
}

func writeJob(chJob chan <- []string, job extractedJob) {
	chJob <- []string{
		"https://www.saramin.co.kr/zf_user/jobs/relay/view?rec_idx=47553391"+job.id, 
		job.title, 
		job.company, 
		job.location, 
		job.experience,
		job.employment,
	}
}

func extractJob(card *goquery.Selection, ch chan <- extractedJob) {
	id, _ := card.Attr("value")
	title := CleanString(card.Find(".area_job>.job_tit>a").Text())
	location := CleanString(card.Find(".area_job>.job_condition>span>a").Text())
	experience := CleanString(card.Find(".area_job>.job_condition>span:nth-child(2)").Text())
	employment := CleanString(card.Find(".area_job>.job_condition>span:nth-child(4)").Text())
	company := CleanString(card.Find(".area_corp > strong > a").Text())
	ch <- extractedJob{
		id: 		id,
		title: 		title,
		location: 	location,
		experience: experience,
		employment: employment,
		company: 	company,
	}
}

func createFile() (w *csv.Writer) {
	file, err := os.Create("jobs.csv")
	checkErr(err)
	
	w = csv.NewWriter(file)

	headers := []string{"Link", "Title", "Company", "Location", "Experience", "Employment"}
	wErr := w.Write(headers)
	checkErr(wErr)

	return
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkStatusCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with Status", res.StatusCode)
	}
}

func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}