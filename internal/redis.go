package news

import (
	"context"
	"fmt"
	"strconv"
	"time"

	redis "gopkg.in/redis.v5"
)

// string value set in redis key news.publish.progress.<zipFileName> when all news of the zip file is published to NEWS_XML
const complete = "complete"

type ErrRedisErrWhileProcessingZipFile struct {
	zipFileName string
	error
}

func (err ErrRedisErrWhileProcessingZipFile) Error() string {
	return fmt.Sprintf("Redis error while processing %s - %s", err.zipFileName, err.error)
}

func publishProgressKey(zipFileName string) string {
	return fmt.Sprintf("news.publish.progress.%s", zipFileName)
}

func publishProgress(redisClient *redis.Client, zipFileName string) (string, error) {
	progress, err := redisClient.Get(publishProgressKey(zipFileName)).Result()
	if err == redis.Nil {
		return "0", nil
	}
	if err != nil {
		return "0", ErrRedisErrWhileProcessingZipFile{zipFileName, err}
	}
	return progress, nil
}

func RedisPublish(ctx context.Context, xmlNewsC <-chan XmlNews, redisClient *redis.Client) {
	for {
		select {
		case <-ctx.Done():
			return
		case xmlNews, ok := <-xmlNewsC:
			if !ok {
				return
			}
			xmls := make([]interface{}, len(xmlNews.xmls))
			for i, pl := range xmlNews.xmls {
				xmls[i] = string(pl)
			}

			for { // iterate until we publish all xmls in this zip file
				progress, err := publishProgress(redisClient, xmlNews.zipFileName)
				if err != nil {
					logger.Println(ErrRedisErrWhileProcessingZipFile{xmlNews.zipFileName, err})
					break
				}
				if progress == complete {
					break
				}
				logger.Printf("number of xmls published for %s=%s\n", xmlNews.zipFileName, progress)
				err = redisClient.Watch(func(tx *redis.Tx) error {
					_, err := tx.Pipelined(func(pipe *redis.Pipeline) error {
						startIndex := 0

						startIndex, err = strconv.Atoi(progress)
						if err != nil {
							return err
						}

						endIndex := startIndex + 66
						if endIndex >= len(xmls) {
							endIndex = len(xmls)
						}

						err = pipe.LPush("NEWS_XML", xmls[startIndex:endIndex]...).Err()
						if err != nil {
							return err
						}
						if endIndex == len(xmls) {
							return pipe.Set(publishProgressKey(xmlNews.zipFileName), complete, 0).Err()
						}
						return pipe.Set(publishProgressKey(xmlNews.zipFileName), endIndex, 0).Err()
					})
					return err
				}, publishProgressKey(xmlNews.zipFileName), "NEWS_XML")
				if err == redis.TxFailedErr {
					logger.Println(err) //this can happen due to the watch's optimistic locking', so it might survive when retried
					time.Sleep(100 * time.Millisecond)
				} else if err != nil {
					logger.Println(ErrRedisErrWhileProcessingZipFile{xmlNews.zipFileName, err})
					break
				}
			}
		}
	}

}
