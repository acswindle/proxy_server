package main

import (
	"errors"
	"net/http"
	"sync"
	"time"
)

type CacheResult struct {
	Header     http.Header
	StatusCode int
	Body       []byte
	Expires    time.Time
}

type Cache map[string]*CacheResult

type CacheMap struct {
	sync.RWMutex
	cache   Cache
	Timeout time.Duration
}

func NewCacheMap(timeout int) *CacheMap {
	result := &CacheMap{
		Timeout: time.Duration(timeout) * time.Second,
		cache:   Cache{},
	}
	go result.cleanCache()
	return result
}

func (c *CacheMap) delete(url string) {
	c.Lock()
	defer c.Unlock()
	delete(c.cache, url)
}

func (c *CacheMap) cleanCache() {
	for range time.Tick(c.Timeout) {
		for key, val := range c.cache {
			if time.Now().After(val.Expires) {
				c.delete(key)
			}
		}
	}
}

func (c *CacheMap) update(url string, result *CacheResult) {
	c.Lock()
	defer c.Unlock()
	c.cache[url] = result
}

func (c *CacheMap) Add(url string, r *CacheResponseWriter) (*CacheResult, error) {
	if c.cache == nil {
		return &CacheResult{}, errors.New("CacheMap is nil")
	}
	result := CacheResult{
		Header: make(http.Header),
	}
	result.StatusCode = r.StatusCode
	result.Header = r.Header()
	result.Body = r.Body.Bytes()
	result.Expires = time.Now().Add(c.Timeout)
	c.update(url, &result)
	return &result, nil
}

func (c *CacheMap) Get(url string) (*CacheResult, error) {
	c.RLock()
	defer c.RUnlock()
	response, present := c.cache[url]
	if !present {
		return nil, errors.New("response not in cache")
	}
	if time.Now().After(response.Expires) {
		go c.delete(url)
		return nil, errors.New("cache result expired")
	}
	return response, nil
}
