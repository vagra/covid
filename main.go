package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
)

func main() {

	tableURL := "https://www.worldometers.info/coronavirus/"
	worldURL := "https://www.worldometers.info/coronavirus/worldwide-graphs/"
	countryURL := "https://www.worldometers.info/coronavirus/country/"

	agent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:74.0) Gecko/20100101 Firefox/74.0"

	country := ""
	printed := false

	sdate := []string{}
	sc := []string{}
	sdc := []string{}
	sd := []string{}
	sdd := []string{}

	// dict := make(map[string]string)

	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exePath := filepath.Dir(exe)

	_, err = os.Stat(filepath.Join(exePath, "data"))
	if os.IsNotExist(err) {
		os.Mkdir(filepath.Join(exePath, "data"), 0777)
	}

	// set input file.
	infile, err := os.Open(filepath.Join(exePath, "country.csv"))
	if err != nil {
		log.Fatal(err)
	}
	defer infile.Close()

	// set trans header file.
	thfile, err := os.Open(filepath.Join(exePath, "trans_header.csv"))
	if err != nil {
		log.Fatal(err)
	}
	defer thfile.Close()

	// set trans country file.
	tcfile, err := os.Open(filepath.Join(exePath, "trans_country.csv"))
	if err != nil {
		log.Fatal(err)
	}
	defer tcfile.Close()

	// read trans_header.csv
	thReader := csv.NewReader(thfile)
	thList, err := thReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	// read trans_country.csv
	tcReader := csv.NewReader(tcfile)
	tcList, err := tcReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	// Instantiate default collector
	c := colly.NewCollector(
		colly.UserAgent(agent),
	)

	c.Limit(&colly.LimitRule{
		RandomDelay: 10 * time.Second,
	})

	// create a request queue with 2 consumer threads
	q, _ := queue.New(
		1, // Number of consumer threads
		&queue.InMemoryQueueStorage{MaxSize: 200000}, // Use default queue storage
	)

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("VISITING:\t", r.URL)
	})

	c.OnResponse(func(r *colly.Response) {
		// fmt.Println("VISITED:\t", r.Request.URL)

		if strings.Contains(r.Request.URL.String(), "worldwide") {
			country = "world"
		} else if strings.Contains(r.Request.URL.String(), "country") {
			country = strings.Replace(r.Request.URL.String(), countryURL, "", -1)
		} else {
			country = "table"
		}

		fmt.Println("VISITED:\t", country)

		printed = false

		sdate = nil
		sc = nil
		sdc = nil
		sd = nil
		sdd = nil

	})

	c.OnHTML("#main_table_countries_yesterday", func(e *colly.HTMLElement) {

		if country != "table" {
			return
		}

		content, err := goquery.OuterHtml(e.DOM)
		fmt.Println("find!")

		outfile, err := os.OpenFile(filepath.Join(exePath, "table.htm"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			panic(err)
		}
		defer outfile.Close()

		writer := bufio.NewWriter(outfile)

		old := `<table id="main_table_countries_yesterday" class="table table-bordered table-hover main_table_countries" style="width:100%;margin-top: 0px !important;display:none;">`

		new := `<table id="main_table_countries_yesterday" class="table table-bordered table-hover main_table_countries dataTable no-footer" style="width:100%;margin-top: 0px !important;">`

		content = strings.Replace(content, old, new, -1)

		for _, trans := range thList {
			src := trans[0]
			dst := trans[1]

			content = strings.Replace(content, src, dst, -1)
		}

		for _, trans := range tcList {
			src := trans[0]
			dst := trans[1]

			content = strings.Replace(content, src, dst, -1)
		}

		writer.WriteString(content)
		writer.Flush()

	})

	c.OnHTML("script", func(e *colly.HTMLElement) {

		content := e.Text

		if country == "table" || printed {
			return
		}

		if strings.Contains(content, `Highcharts.chart('coronavirus-cases-linear'`) ||
			strings.Contains(content, `Highcharts.chart('coronavirus_cases_daily'`) ||
			strings.Contains(content, `Highcharts.chart('coronavirus-deaths-linear'`) ||
			strings.Contains(content, `Highcharts.chart('coronavirus-deaths-daily'`) ||
			strings.Contains(content, `Highcharts.chart('graph-cases-daily'`) ||
			strings.Contains(content, `Highcharts.chart('graph-deaths-daily'`) {

			regx := regexp.MustCompile(`(?s)categories:\s*\[(.*?)\]`)
			date := regx.FindStringSubmatch(content)[1]
			sdate = strings.Split(date, ",")

			regy := regexp.MustCompile(`(?s)series:\s*\[\s*\{(.*?)]\s*\}`)
			series := regy.FindString(content)

			regn := regexp.MustCompile(`(?s)name:\s*'(.*?)'`)
			regd := regexp.MustCompile(`(?s)data:\s*\[(.*?)\]`)
			name := regn.FindStringSubmatch(series)[1]
			data := regd.FindStringSubmatch(series)[1]

			switch name {
			case "Cases":
				sc = strings.Split(data, ",")
			case "Daily Cases":
				sdc = strings.Split(data, ",")
			case "Deaths":
				sd = strings.Split(data, ",")
			case "Daily Deaths":
				sdd = strings.Split(data, ",")
			}
		}

		if len(sdate)*len(sc)*len(sdc)*len(sd)*len(sdd) > 0 {
			log.Printf("%d, %d, %d, %d, %d", len(sdate), len(sc), len(sdc), len(sd), len(sdd))

			outfile, err := os.OpenFile(filepath.Join(exePath, "data/"+country+".csv"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
			if err != nil {
				panic(err)
			}
			defer outfile.Close()

			writer := bufio.NewWriter(outfile)

			writer.WriteString("Date,Cases,Daily Cases,Death,Daily Deaths\n")

			for i := 0; i < len(sdate); i++ {
				writer.WriteString(sdate[i] + "," + sc[i] + "," + sdc[i] + "," + sd[i] + "," + sdd[i] + "\n")
			}

			writer.Flush()

			printed = true
		}

	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("ERROR:\t", r.Request.URL)
		fmt.Println(err)
		r.Request.Retry()
	})

	q.AddURL(tableURL)
	q.AddURL(worldURL)

	// read country.csv
	csvReader := csv.NewReader(infile)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	for _, record := range records {
		url := record[0]
		// name := record[1]
		// dict[url] = name

		if url == `country` ||
			url == `world` {
			continue
		}

		q.AddURL(countryURL + url)
	}

	scanner := bufio.NewScanner(infile)
	for scanner.Scan() {
		line := scanner.Text()

		if line == `country` ||
			line == `world` {
			continue
		}

		q.AddURL(countryURL + line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	q.Run(c)
	log.Println(c)
}
