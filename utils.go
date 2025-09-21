package main

import "net/http"

func copyHeader(w http.Header, o http.Header) {
	for key, vals := range o {
		for _, val := range vals {
			w.Add(key, val)
		}
	}
}
