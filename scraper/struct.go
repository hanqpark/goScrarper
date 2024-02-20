package scraper

type extractedJob struct {
	id 	  	 	string
	title 	 	string
	company 	string
	location 	string
	experience	string
	employment  string
}

var (
	baseURL string = "https://www.saramin.co.kr/zf_user/search/recruit?&recruitPageCount=40"
	FILE_NAME string = "jobs.csv"
)