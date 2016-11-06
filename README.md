# About the proj

This is only a small project done as part of an interview process.

This URL (http://bitly.com/nuvi-plz) is an http folder containing a list of zip files. Each zip file contains a bunch of xml files. Each xml file contains 1 news report.
This application needs to download all of the zip files, extract out the xml files, and publish the content of each xml file to a redis list called “NEWS_XML”.

Runs should be idempotent and runnable multiple times, but without submitting duplicate data to the redis list.

## Requirements

- Redis
- Go
- Git

## Download

```shell
go get github.com/xushichangdesmond/news
```

## Using docker to run Redis

```shell
docker run --name redis -d -p6379:6379 redis redis-server --requirepass test
```

## Running

```shell
go run news.go
```

By default it connects to redis running at localhost:6379 using 'test' as password
To configure these differently (and other parameters), see the command options available via
```shell
go run news.go -h
```