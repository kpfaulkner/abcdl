package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"regexp"
)

const (
	abcPrefix string = "iview.abc.net.au/video/ZW"
	urlRegex string = "iview.abc.net.au/video/Z[a-zA-Z0-9]*"
)


func downloadContentsfromURL( url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

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

	checker := regexp.MustCompile(urlRegex)
	doesMatch := checker.MatchString( contents)
	if doesMatch {
		urlList = checker.FindAllString(contents,100)
	}

	// unique.
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
	
}
