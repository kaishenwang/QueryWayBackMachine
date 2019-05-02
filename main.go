package main

import (
	"sync"
	"os"
	"log"
	"encoding/csv"
	"bufio"
	"io"
	"fmt"
	"time"
	"strconv"
	"net/url"
	"github.com/GiterLab/urllib"
)

func generateURL(wg *sync.WaitGroup, output chan<- string) {
	defer (*wg).Done()

	csvFile, _ := os.Open("/Users/kaishenwang/CT/blacklist-web-measurement/data/new_bl_domains_04252019_to_05012019.csv")
	reader := csv.NewReader(bufio.NewReader(csvFile))
	count := 0
	for {
		count += 1
		if count > 2000 {
			break
		}
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
		encodeDomain := url.QueryEscape(line[0])
		startDate := convertDateTime(line[1], true)
		endDate := convertDateTime(line[1], false)
		output <- fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=%s&from=%s&to=%s",
			encodeDomain, startDate,endDate)

	}
}

func makeQuery(wg *sync.WaitGroup, input <-chan string) {
	defer (*wg).Done()
	for query := range(input) {
		//fmt.Println(query)
		res,err := urllib.Get(query).String()

		//resp, err := http.Get(query)
		if err != nil {
			log.Fatal(err)
		}
		/*
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		res := buf.String()
		*/
		if len(res) < 5 {
			continue
		}
		fmt.Println("------------")
		fmt.Print(res)
		fmt.Println("------------")

	}

}

func main() {
	workerCount := 1
	urlChan := make (chan string)
	outputChan := make (chan string)
	var generateUrlWG, queryWG sync.WaitGroup
	generateUrlWG.Add(1)
	queryWG.Add(workerCount)
	go generateURL(&generateUrlWG, urlChan)
	for i := 0; i < workerCount; i++ {
		go makeQuery(&queryWG, urlChan)
	}
	generateUrlWG.Wait()
	close(urlChan)
	queryWG.Wait()
	close(outputChan)

}



// Utility
func convertDateTime(t string, oneMonthAgo bool) string {
	i, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		panic(err)
	}
	tm := time.Unix(i, 0)
	if oneMonthAgo {
		tm = tm.AddDate(0,0,-30)
	}
	return fmt.Sprintf("%d%d%d%d%d%d", tm.Year(), tm.Month(), tm.Day(),
		tm.Hour(), tm.Minute(), tm.Second())
}