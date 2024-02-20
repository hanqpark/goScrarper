package main

import (
	"os"
	"strings"

	"github.com/hanqpark/goScraper/scraper"
	"github.com/labstack/echo"
)

func handleHome(c echo.Context) error {
	return c.File("index.html")
}

func handleScrape(c echo.Context) error {
	defer os.Remove(scraper.FILE_NAME)
	term := strings.ToLower(scraper.CleanString(c.FormValue("term")))
	scraper.Scrape(term)
	return c.Attachment(scraper.FILE_NAME, term + ".csv")
}

func main() {
	e := echo.New()
	e.GET("/", handleHome)
	e.POST("/scrape", handleScrape)
	e.Logger.Fatal(e.Start(":1323"))
}

