package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/tamalsaha/prometheus-demo/prometheus"
	"sigs.k8s.io/yaml"
)

// ref: https://kubernetes.io/docs/tasks/administer-cluster/access-cluster-services/#manually-constructing-apiserver-proxy-urls
func main() {
	cfg := prometheus.Config{
		Addr: "http://localhost:9090",
		// BasicAuth:       prometheus.BasicAuth{},
		// BearerToken:     "",
		// BearerTokenFile: "",
		// ProxyURL:        "",
		// TLSConfig:       prom_config.TLSConfig{},
	}
	pc, err := cfg.NewPrometheusClient()
	if err != nil {
		panic(err)
	}
	pc2 := promv1.NewAPI(pc)

	promCPUQuery := `up`

	res, err := getPromQueryResult(pc2, promCPUQuery)
	if err != nil {
		log.Fatalf("failed to get prometheus cpu query result, reason: %v", err)
	}
	data, _ := yaml.Marshal(res)
	fmt.Println(string(data))
}

func getPromQueryResult(pc promv1.API, promQuery string) (map[string]float64, error) {
	val, warn, err := pc.Query(context.Background(), promQuery, time.Now())
	if err != nil {
		return nil, err
	}
	if warn != nil {
		log.Println("Warning: ", warn)
	}

	metrics := strings.Split(val.String(), "\n")

	cpu := float64(0)

	metricsMap := make(map[string]float64)

	for _, m := range metrics {
		val := strings.Split(m, "=>")
		if len(val) != 2 {
			return nil, fmt.Errorf("metrics %q is invalid for query %s", m, promQuery)
		}
		valStr := strings.Split(val[1], "@")
		if len(valStr) != 2 {
			return nil, fmt.Errorf("metrics %q is invalid for query %s", m, promQuery)
		}
		valStr[0] = strings.Replace(valStr[0], " ", "", -1)
		metricVal, err := strconv.ParseFloat(valStr[0], 64)
		if err != nil {
			return nil, err
		}
		cpu += metricVal

		metricsMap[val[0]] = metricVal
	}

	return metricsMap, nil
}
