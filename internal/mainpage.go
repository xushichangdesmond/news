package news

import (
	"fmt"
	"io"
	"net/http"

	"strings"

	"net/url"

	"bytes"

	"io/ioutil"

	"golang.org/x/net/html"
)

type HttpError int

func (err HttpError) Error() string {
	return fmt.Sprintf("%d - %s", err, http.StatusText(int(err)))
}

func ExtractZipFileNames(startUrl string) (baseLoc *url.URL, zipFileNames <-chan string, err error) {
	resp, err := http.Get(startUrl)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, HttpError(resp.StatusCode)
	}

	// we don't want to keep the req/resp open for longer than necessary, so we will just read it all in first
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	return resp.Request.URL, extractZipFileNames(bytes.NewReader(body)), nil
}

//TODO: randomize order of returned filenames - so that we can more easily run multiple instances (from different machines) to help finish stuff faster
func extractZipFileNames(body io.Reader) <-chan string {
	z := html.NewTokenizer(body)
	results := make(chan string)

	go func() {
		for {
			tt := z.Next()
			switch tt {
			case html.ErrorToken:
				if z.Err() == io.EOF {
					close(results)
				}
				logger.Printf("Error parsing - %s", z.Err()) // TODO: consider if we need better error handling here
				return
			case html.StartTagToken:
				t := z.Token()
				if t.Data != "a" {
					continue
				}
				for _, att := range t.Attr {
					if att.Key == "href" {
						if strings.HasSuffix(att.Val, ".zip") {
							results <- att.Val
						}
						break

					}
				}
			}
		}
	}()
	return results
}
