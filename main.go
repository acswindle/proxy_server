package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"time"
)

type ProxyConfig struct {
	port          int
	url           string
	clientTimeout int
	cacheTimeout  int
}

func ParseFlags() *ProxyConfig {
	config := ProxyConfig{}
	flag.IntVar(&config.port, "port", 9999, "port to serve proxy on")
	flag.StringVar(&config.url, "url", "http://httpbin.org", "url to forward to")
	flag.IntVar(&config.clientTimeout, "ctime", 10, "client timeout time in seconds")
	flag.IntVar(&config.cacheTimeout, "cachetime", 5, "time to cache results in minutes")
	flag.Parse()
	return &config
}

type ProxyServer struct {
	config *ProxyConfig
	client http.Client
	logger *slog.Logger
	cache  *CacheMap
}

func (s ProxyServer) ForwardRequest(r *http.Request) *http.Response {
	fullPath := fmt.Sprintf("%s%s%s", s.config.url, r.URL.Path, r.URL.RawQuery)
	request, err := http.NewRequest(r.Method, fullPath, r.Body)
	request.Header = r.Header
	if err != nil {
		panic(err)
	}
	response, err := s.client.Do(request)
	if err != nil {
		panic(err)
	}
	return response
}

func CopyResponse(ClientResponse *http.Response, w http.ResponseWriter) {
	copyHeader(w.Header(), ClientResponse.Header)
	w.WriteHeader(ClientResponse.StatusCode)
	defer ClientResponse.Body.Close()
	io.Copy(w, ClientResponse.Body)
}

func (s ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := s.ForwardRequest(r)
	CopyResponse(response, w)
}

func main() {
	config := ParseFlags()
	server := ProxyServer{
		config,
		http.Client{
			Timeout: time.Duration(config.clientTimeout) * time.Second,
		},
		slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelInfo,
		})),
		NewCacheMap(config.cacheTimeout),
	}
	server.logger.Info("Starting Proxy Server", "port", config.port,
		"forward-url", config.url)
	if err := http.ListenAndServe(
		fmt.Sprintf(":%d", config.port),
		server.CacheMiddlware(server),
	); err != nil {
		server.logger.Error("could not start http listener", "error", err,
			"stack", debug.Stack())
		os.Exit(1)
	}
}
