package main

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
)

type rt struct {
}

func (r rt) RoundTrip(req *http.Request) (*http.Response, error) {
	r2 := req.Clone(context.TODO())
	r2.Host = ""
	r2.URL.Path = path.Join("/trickster-auth/", r2.URL.Path)
	r2.Method = http.MethodGet
	r2.Body = nil
	resp, err := http.DefaultTransport.RoundTrip(r2)
	if err != nil || resp.StatusCode != http.StatusOK {
		return resp, err
	}

	// forward to trickster

	// prometheus.appscode.ninja/1-****/
	// /trickster-auth/{$id}/*

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
