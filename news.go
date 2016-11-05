package main

import (
	"flag"
	"log"
	"os"

	redis "gopkg.in/redis.v5"

	"sync"

	"context"

	"net/http"
	_ "net/http/pprof"

	news "github.com/xushichangdesmond/news/internal"
)

var (
	startURL               = flag.String("startURL", "http://bitly.com/nuvi-plz", "url of http folder to find zips containing the xml news")
	reportDownloadProgress = flag.Bool("reportDownloadProgress", true, "Whether download progress of zip files to be reported in logs")
	redisURL               = flag.String("redisURL", "localhost:6379", "host:port of redis instance")
	redisPassword          = flag.String("redisPassword", "test", "password to login to redis instance")
	pprofURL               = flag.String("pprofURL", "localhost:6060", "host:port to run pprofURL on, or blank to turn it off")
)

var logger = log.New(os.Stderr, "news - ", log.LstdFlags)

func main() {
	flag.Parse()

	if *pprofURL != "" {
		go func() {
			http.ListenAndServe(*pprofURL, nil)
		}()
	}

	baseLoc, zips, err := news.ExtractZipFileNames(*startURL)
	if err != nil {
		logger.Fatal(err)
	}

	var downloadWorkers, redisPublishingWorkers sync.WaitGroup
	ctx := context.Background() // not really used now, but we want to support for the future

	xmlNewsC := make(chan news.XmlNews, 5)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     *redisURL,
		Password: *redisPassword,
	})
	defer redisClient.Close()

	// 3 download workers
	for i := 0; i < 1; i++ {
		downloadWorkers.Add(1)
		go func() {
			defer downloadWorkers.Done()
			news.Download(ctx, zips, baseLoc, redisClient, xmlNewsC, *reportDownloadProgress)
		}()
	}

	// 1 redis publishing worker - since we will be writing into the same NEWS_XML key,
	// it is likely that having more publishing goroutines only increases the chance for the optimistic locking to fail (and require retries)
	redisPublishingWorkers.Add(1)
	go func() {
		defer redisPublishingWorkers.Done()
		news.RedisPublish(ctx, xmlNewsC, redisClient)
	}()

	downloadWorkers.Wait()
	close(xmlNewsC)
	redisPublishingWorkers.Wait()

	logger.Println("Thats all folks")
}
