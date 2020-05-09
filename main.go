package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"time"
)

const (
	//abcPrefix string = "iview.abc.net.au/video/ZW"
	//abcCDNPrefix string = "cdn.iview.abc.net.au/thumbs/.+/dr/(DR[a-zA-Z0-9]*)/._.*"
	urlRegex string = "iview.abc.net.au/video/Z[a-zA-Z0-9]*"
	showRegex string ="https://iview.abc.net.au/show/(.+)"
	urlTemplate string = "https://iview.abc.net.au/video/%s"

	// just an example
	// %s is show name
	// %d is page/series (unsure which)  either way 1,2,3.....
	apiURLFormat string ="https://api.iview.abc.net.au/v2/series/%s/%d"
)

type IViewSeriesSubResponse struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	ShowTitle       string `json:"showTitle"`
	Description     string `json:"description"`
	DisplayTitle    string `json:"displayTitle"`
	DisplaySubtitle string `json:"displaySubtitle"`
	Thumbnail       string `json:"thumbnail"`

	Embedded struct {
		VideoEpisodes []struct {
			ID                        string   `json:"id"`
			Channel                   string   `json:"channel"`
			ChannelTitle              string   `json:"channelTitle"`
			Type                      string   `json:"type"`
			HouseNumber               string   `json:"houseNumber"`
			Title                     string   `json:"title"`
			ShowTitle                 string   `json:"showTitle"`
			SeriesTitle               string   `json:"seriesTitle"`
			DisplayTitle              string   `json:"displayTitle"`
			DisplaySubtitle           string   `json:"displaySubtitle"`
			Thumbnail                 string   `json:"thumbnail"`
			Description               string   `json:"description"`
			PubDate                   string   `json:"pubDate"`
			ExpireDate                string   `json:"expireDate"`
			Tags                      []string `json:"tags"`
			Duration                  int      `json:"duration"`
			DisplayDuration           string   `json:"displayDuration"`
			DisplayDurationAccessible string   `json:"displayDurationAccessible"`
			Classification            string   `json:"classification"`
			Captions                  bool     `json:"captions"`
			CaptionsOnAkamai          bool     `json:"captionsOnAkamai"`
			Availability              string   `json:"availability"`
			Participants              []struct {
				Title string `json:"title"`
				List  string `json:"list"`
			} `json:"participants,omitempty"`
		} `json:"videoEpisodes"`
	} `json:"_embedded"`
}



func downloadContentsfromURL( url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// bail if not 200 (probably 404)
	if resp.StatusCode != 200 {
		return "", errors.New("Unable to download contents")
	}

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	s := string(byteArray)

	return s, nil
}

// getURLsfromPage get urls from page. Return as list of strings.
// urls have to have prefix of abcPrefix
func getURLsfromPage( url string ) ([]string, error) {

	urlList := []string{}
	contents, err := downloadContentsfromURL(url)
	if err != nil {
		return nil, err
	}

	// regular.
	checker := regexp.MustCompile(urlRegex)
	doesMatch := checker.MatchString( contents)
	if doesMatch {
		urlList = checker.FindAllString(contents,100)
	}
	uniqueList := uniqueStringList(urlList)


  return uniqueList, nil
}

// getURLsfromPageWithName get urls from page. Return as list of strings.
// urls have to have prefix of abcPrefix
func getURLsfromPageWithName( name string ) ([]string, error) {

	urlList := []string{}

	index := 1
	for {
		url := fmt.Sprintf(apiURLFormat, name, index)
		contents, err := downloadContentsfromURL(url)
		if err != nil {
			fmt.Printf("Unable to download contents %s\n", err.Error())
			//return nil, err
			break
		}

		var resp IViewSeriesSubResponse
		err = json.Unmarshal([]byte(contents), &resp)
		if err != nil {
			fmt.Printf("Unable to unmarshal contents %s\n", err.Error())
			break
		}

		for _, ep := range resp.Embedded.VideoEpisodes {
			newUrl := fmt.Sprintf(urlTemplate, ep.HouseNumber)
			urlList = append(urlList, newUrl)
		}

		// next page of links.
		index++
	}

	uniqueList := uniqueStringList(urlList)

	return uniqueList, nil
}

func uniqueStringList( l []string) []string {
	m := make(map[string]bool)
  uniqueList := []string{}
	for _,s := range l {
		if _,ok := m[s] ; !ok  {
			uniqueList = append(uniqueList, s)
			m[s] = true
		}
	}
	return uniqueList
}

func main() {
	fmt.Printf("so it begins....\n")

	url := flag.String("url", "", "url")
	name := flag.String("name", "", "name of last part of URL path. eg. https://iview.abc.net.au/show/<THIS BIT>. This will recursively get a LOT of shows off the page.")
	fileName := flag.String("file","","file with urls")

	flag.Parse()
	if *url != "" {
		l, _ := getURLsfromPage(*url)

		// download them all.
		for _, url := range l {
			fmt.Printf("Downloading %s\n", url)
			c := exec.Command("youtube-dl", "-v", url)
			if err := c.Run(); err != nil {
				fmt.Println("Error: ", err)
			}
		}
	}

	if *name != "" {
		l, _ := getURLsfromPageWithName(*name)

		// download them all.
		for _, url := range l {
			time.Sleep( time.Duration(rand.Intn(60)) * time.Second)
			fmt.Printf("Downloading %s\n", url)
			c := exec.Command("youtube-dl", "-v", url)
			if err := c.Run(); err != nil {
				fmt.Println("Error: ", err)
			}
		}
	}

	if *fileName != "" {
		// read file and split on newline.
		file, err := os.Open( *fileName)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			f := scanner.Text()
			if f != "" {

				// random sleep... just incase monitoring for too quick turn arounds
				time.Sleep( time.Duration(rand.Intn(60)) * time.Second)
				fmt.Printf("Downloading %s\n", f)
				c := exec.Command("youtube-dl", "-v", f)
				if err := c.Run(); err != nil {
					fmt.Println("Error: ", err)
				}
			}
		}
	}

}
