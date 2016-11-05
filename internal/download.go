package news

import (
	"context"
	"net/url"

	redis "gopkg.in/redis.v5"
)

func Download(ctx context.Context, zipFileNames <-chan string, baseLoc *url.URL, redisClient *redis.Client, results chan<- XmlNews, reportDownloadProgress bool) {
	for {
		select {
		case <-ctx.Done():
			return
		case zipFileName, ok := <-zipFileNames:
			if !ok {
				return
			}
			progress, err := publishProgress(redisClient, zipFileName)
			if err != nil {
				logger.Println(ErrRedisErrWhileProcessingZipFile{zipFileName, err})
				continue
			}
			if progress == "complete" {
				logger.Printf("zip file already processed before %s\n", zipFileName)
				continue
			}

			logger.Printf("going to download %s\n", zipFileName)
			zipURL, err := url.Parse(zipFileName)
			if err != nil {
				logger.Printf("Error - Invalid zip link %s - skipping - %s\n", zipFileName, err)
				continue
			}

			zipURL = baseLoc.ResolveReference(zipURL)
			logger.Printf("zipURL %s\n", zipURL.String())

			xmls, err := DownloadAndExtractFromZip(zipURL, reportDownloadProgress)
			if err != nil {
				logger.Println(err)
			}
			logger.Printf("payloads extracted - %s\n", zipURL.String())

			results <- XmlNews{zipFileName, xmls}
		}
	}
}
