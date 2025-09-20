package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
)

type ProxyConfig struct {
	port int
	url  string
}

func ParseFlags() *ProxyConfig {
	config := ProxyConfig{}
	flag.IntVar(&config.port, "port", 9999, "port to serve proxy on")
	flag.StringVar(&config.url, "url", "http://httpbin.org", "url to forward to")
	flag.Parse()
	return &config
}

type ProxyServer struct {
	config *ProxyConfig
	client http.Client
	logger *slog.Logger
}

func (s ProxyServer) ForwardRequest(subpath string, method string) *http.Response {
	buffer := bytes.NewBuffer([]byte{})
	fullPath := fmt.Sprintf("%s%s", s.config.url, subpath)
	request, err := http.NewRequest(method, fullPath, buffer)
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
	for key, vals := range ClientResponse.Header {
		for _, val := range vals {
			w.Header().Add(key, val)
		}
	}
	w.Header().Add("X-Cache", "MISS")
	w.WriteHeader(ClientResponse.StatusCode)
	io.Copy(w, ClientResponse.Body)
}

func (s ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("Forwarding request", "subpath", r.URL.Path, "method", r.Method)
	response := s.ForwardRequest(r.URL.Path, r.Method)
	CopyResponse(response, w)
}

func main() {
	config := ParseFlags()
	server := ProxyServer{
		config,
		http.Client{},
		slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelInfo,
		})),
	}
	server.logger.Info("Starting Proxy Server", "port", config.port,
		"forward-url", config.url)
	if err := http.ListenAndServe(
		fmt.Sprintf(":%d", config.port),
		server,
	); err != nil {
		server.logger.Error("could not start http listener", "error", err,
			"stack", debug.Stack())
		os.Exit(1)
	}
}
