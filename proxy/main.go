package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type rt struct {
}

func (r rt) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Host = ""
	if data, err := httputil.DumpRequest(req, true); err == nil {
		fmt.Println(string(data))
	}
	return http.DefaultTransport.RoundTrip(req)
}

var _ http.RoundTripper = &rt{}

func main() {
	rpURL, err := url.Parse("https://trickster.appscode.ninja")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("http://localhost:3000")

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	proxy := httputil.NewSingleHostReverseProxy(rpURL)
	proxy.Transport = &rt{}

	r.Handle("/*", proxy)
	http.ListenAndServe(":3000", r)

	select {}

	//resp, err := http.Get(frontendProxy.URL)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//b, err := io.ReadAll(resp.Body)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//fmt.Printf("%s", b)
}
