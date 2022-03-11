package main

import (
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/tamalsaha/prometheus-demo/prometheus"
)

func main() {
	var cfg prometheus.Config
	pc, err := cfg.NewPrometheusClient()
	if err != nil {
		panic(err)
	}
	pc2 := promv1.NewAPI(pc)
}
