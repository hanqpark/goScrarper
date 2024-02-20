package scrape

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

var (
	baseURL string = "https://www.saramin.co.kr/zf_user/search/recruit?&searchword=python&recruitPageCount=40"
	userAgent string = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"
)

func GetPage(page int, chMain chan <- []ExtractedJob) {
	pageURL := baseURL + "&recruitPage=" + strconv.Itoa(page)
	fmt.Println(pageURL)

	res, err := http.Get(pageURL)
	checkErr(err)
	checkStatusCode(res)
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	ch := make(chan ExtractedJob)
	searchCards := doc.Find(".item_recruit")
	searchCards.Each(func(i int, card *goquery.Selection){
		go extractJob(card, ch)
	})

	var jobs []ExtractedJob
	for i:=0; i<searchCards.Length(); i++ {
		job := <- ch
		jobs = append(jobs, job)
	}
	chMain <- jobs
}

func GetPages() (pages int) {
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

func WriteJobs(jobs []ExtractedJob, w *csv.Writer, wg *sync.WaitGroup) {
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

func writeJob(chJob chan <- []string, job ExtractedJob) {
	chJob <- []string{
		"https://www.saramin.co.kr/zf_user/jobs/relay/view?rec_idx=47553391"+job.id, 
		job.title, 
		job.company, 
		job.location, 
		job.experience,
		job.employment,
	}
}

func extractJob(card *goquery.Selection, ch chan <- ExtractedJob) {
	id, _ := card.Attr("value")
	title := cleanString(card.Find(".area_job>.job_tit>a").Text())
	location := cleanString(card.Find(".area_job>.job_condition>span>a").Text())
	experience := cleanString(card.Find(".area_job>.job_condition>span:nth-child(2)").Text())
	employment := cleanString(card.Find(".area_job>.job_condition>span:nth-child(4)").Text())
	company := cleanString(card.Find(".area_corp > strong > a").Text())
	ch <- ExtractedJob{
		id: 		id,
		title: 		title,
		location: 	location,
		experience: experience,
		employment: employment,
		company: 	company,
	}
}

func CreateFile() (w *csv.Writer) {
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

func cleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}