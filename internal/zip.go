package news

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

var logger = log.New(os.Stderr, "news - ", log.LstdFlags)

type ErrRetrievingZip struct {
	*url.URL
	error
}

func (err ErrRetrievingZip) Error() string {
	return fmt.Sprintf("Error retrieving zip at %s - %s", err.URL.String(), err.error.Error())
}

type ErrProcessingZip struct {
	*url.URL
	*zip.File
	error
}

func (err ErrProcessingZip) Error() string {
	return fmt.Sprintf("Error processing file %s in zip at %s - %s", err.File.Name, err.URL.String(), err.error.Error())
}

// dev notes - not using channel for streaming model only because we need to read in zip fully before processing it (it needs random access)
func DownloadAndExtractFromZip(zipUrl *url.URL, reportDownloadProgress bool) ([][]byte, error) {
	resp, err := http.Get(zipUrl.String())
	if err != nil {
		return nil, ErrRetrievingZip{zipUrl, err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrRetrievingZip{zipUrl, HttpError(resp.StatusCode)}
	}

	/*payload, err := getBytes(resp.Body, zipUrl, resp.ContentLength, reportDownloadProgress)
	if err != nil {
		return nil, ErrRetrievingZip{zipUrl, err}
	}*/

	payload, err := ioutil.ReadFile("f:/1478077143146.zip")
	if err != nil {
		return nil, ErrRetrievingZip{zipUrl, err}
	}

	reader, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	if err != nil {
		return nil, ErrRetrievingZip{zipUrl, err}
	}

	pl := make([][]byte, len(reader.File))
	for i, f := range reader.File {
		pl[i], err = extractFileInZip(f)
		if err != nil {
			return nil, ErrProcessingZip{zipUrl, f, err}
		}
	}
	return pl[:356], nil
}

func extractFileInZip(f *zip.File) ([]byte, error) {
	r, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return ioutil.ReadAll(r)
}

func getBytes(reader io.Reader, zipUrl *url.URL, contentLength int64, reportDownloadProgress bool) ([]byte, error) {
	if reportDownloadProgress {
		b := make([]byte, 1<<13)
		buf := bytes.Buffer{}
		for {
			n, err := reader.Read(b)
			if err == io.EOF {
				_, err = buf.Write(b[:n])
				if err != nil {
					logger.Fatal(err)
				}
				break
			} else if err != nil {
				return nil, err
			}
			_, err = buf.Write(b[:n])
			if err != nil {
				return nil, err
			}
			logger.Printf("zipUrl - %s; %d/%d\n", zipUrl, buf.Len(), contentLength)
		}
		return buf.Bytes(), nil
	}
	return ioutil.ReadAll(reader)

}
