# About the proj

This is only a small project done as part of an interview process.

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
docker run --name redis -d redis
```

## Running

```shell
go run news.go
```

For command line options and setting various params like redis url, look at
```shell
go run news.go -h
```