# Caching Proxy

This is a basic implementation for a simple caching proxy server. It roughly follows the guidelines
set out in [this roadmap.sh project](https://roadmap.sh/projects/caching-server).

## Goals

The goals I had undergoing this project were to practice

1. parsing cli arguments using Go's flag package
2. using Go's http client for forwarding requests
3. creating a customer proxy server that implements the net.Handler interface
4. create a simple in-memory cache that utilizes good concurrency best practices such as sync.Lock
5. creating a cache middleware
6. create a background process using go functions and channels to clean up the cache

## Project Layout

```
project
| main.go -- Entry point to server, where flags are parsed, and forward proxy logic implemented
| utils.go -- Helper functions, currently just a function to shallow copy http headers
| cache.go -- Cache data structure with access methods
| middleware.go -- Where the cache proxy logic is implemented
```

## Design Choices

I decided to implement the cache mechanism as middleware instead of in the ServeHTTP method. I chose this because

1. This server could become a more feature full proxy server aside from just caching. Making caching a middleware
   allows for additional features to be added on later without major refactoring of the caching / forwarding code
2. I wanted to play around with a middleware that needed to read back the body a response of the inner handlers.

For the logic of the Cache itself, I decided to only implement it for GET requests. I chose not to read the
response body for headers like cache control, etc. in the interest of time. In a true production environment,
I would spend time making the logic [RFC 7234](https://datatracker.ietf.org/doc/html/rfc7234) complaint.

I also did not implement clearing the cache as mentioned on the roadmap.sh prompt. Since I used an in process and in memory cache,
I wasn't sure how to implement it as they suggested. I might in the future implementing it by
inspecting the request header for basic auth credentials and a reset header, but that's a lot of work for a simple toy.

## Usage

To run the server, first clone the repo and then run

```bash
go run . --url URLTOFORWARDTO --port PORT --ctime CLIENTTIMEOUT --cachetimeout CACHETIMEOUT
```
