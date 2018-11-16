package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/anaskhan96/soup"
)

type Tweet struct {
	Content, Tweet string
	Urls           []string
}

const (
	usage = "Usage:\n go-twitter-to-es -csv=Path/to/CSV\n Takes a IFTTT 'Liked Tweeets' csv and imports the tweets into an Elasticsearch, 4th field is the link to the tweet."
	esURL = "http://localhost:9200/twitter/tweet"
)

func main() {
	// process command line options
	var location string
	helpPtr := flag.Bool("help", false, "display help")
	hPtr := flag.Bool("h", false, "display help")
	flag.StringVar(&location, "location", "", "the Path to the csv file")
	flag.Parse()
	if *hPtr || *helpPtr {
		fmt.Println(usage)
		os.Exit(0)
	}
	if location == "" {
		fmt.Println(usage)
		os.Exit(1)
	}

	// pass location to the parse function
	data, err := parseCsv(location)
	if err != nil {
		os.Exit(1)
	}

	for _, element := range data {
		//empty lines in csv happpen
		if len(element) == 0 {
			continue
		}
		resp, err := soup.Get(element)
		if err != nil {

			os.Exit(1)
		}
		var content string
		var url []string
		doc := soup.HTMLParse(resp)
		twID := strings.Split(element, "/")[5]
		dataTweet := doc.FindAll("div", "data-tweet-id", twID)
		// sometimes there is a nullpoint dereference error when using Find()
		// on the above line, this is my workaround
		for _, data := range dataTweet {
			p := data.FindAll("p")
			for _, cont := range p {
				content += cont.Text()
			}
			// get urls for external content, this goes through the twitter
			// url shortener. needs some processing to get the original url
			urlCont := data.FindAll("a")
			for _, container := range urlCont {
				keys := container.Attrs()
				if strings.Contains(keys["href"], "https") {
					url = append(url, keys["href"])
				}
			}
		}
		// create tweet object for seralization later on
		t := Tweet{}
		t.Content = strings.TrimSpace(content)
		t.Urls = url
		t.Tweet = element
		// create json object
		b, err := json.Marshal(t)
		if err != nil {
			fmt.Printf("Error: %s", err)
			return
		}
		// post to ElasticSearch for indexing
		postToES(b)
	}
}

// taken from stackoverflow. thanks internet
func parseCsv(filename string) ([]string, error) {
	a := make([]string, 1)
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	csvr := csv.NewReader(file)

	for {
		row, err := csvr.Read()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return a, err
		}
		a = append(a, row[3])
	}
}

//pass json.Marshal data to this function
func postToES(tweet []byte) {
	payload := strings.NewReader(string(tweet))

	req, _ := http.NewRequest("POST", esURL, payload)
	// content-type needed for ES
	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	// info if the indexing succeeded
	fmt.Println(string(body))

}
