package main

import (
	"fmt"
	"net/http"

	"github.com/MarkRepo/Gee/Gee/gee"
)

func main() {
	r := gee.New()

	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		_, _ = fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
	})

	r.Get("/hello", func(w http.ResponseWriter, req *http.Request) {
		for k, v := range req.Header {
			_, _ = fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
		}
	})

	_ = r.Run(":9999")

}
