package main

import (
	"context"
	"encoding/json"
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	kutil "kmodules.xyz/client-go"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	prom_config "github.com/prometheus/common/config"

	"k8s.io/client-go/rest"

	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/tamalsaha/prometheus-demo/prometheus"
	ctrl "sigs.k8s.io/controller-runtime"
)

var tmpDir = func() string {
	dir, err := os.MkdirTemp("/tmp", "prometheus-*")
	if err != nil {
		panic(err)
	}
	return dir
}()

type ServiceReference struct {
	Scheme    string
	Name      string
	Namespace string
	Port      int
}

func ToPrometheusConfigFromServiceAccount(cfg *rest.Config, sa types.NamespacedName, ref ServiceReference) (*prometheus.Config, error) {
	kc, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	var caData, tokenData []byte
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (done bool, err error) {
		secret, err := kc.CoreV1().Secrets(sa.Namespace).Get(context.TODO(), sa.Name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return false, nil
		} else if err != nil {
			return false, err
		}

		var caFound, tokenFound bool
		caData, caFound = secret.Data["ca.crt"]
		tokenData, tokenFound = secret.Data["token"]
		return caFound && tokenFound, nil
	})
	if err != nil {
		return nil, err
	}

	certDir, err := os.MkdirTemp(os.TempDir(), "prometheus-*")
	if err != nil {
		return nil, err
	}

	caFile := filepath.Join(certDir, "ca.crt")
	if err := os.WriteFile(caFile, caData, 0o644); err != nil {
		return nil, err
	}

	return &prometheus.Config{
		Addr:        fmt.Sprintf("%s/api/v1/namespaces/%s/services/%s:%s:%d/proxy/", cfg.Host, ref.Namespace, ref.Scheme, ref.Name, ref.Port),
		BearerToken: string(tokenData),
		ProxyURL:    "",
		TLSConfig: prom_config.TLSConfig{
			CAFile:             caFile,
			ServerName:         cfg.TLSClientConfig.ServerName,
			InsecureSkipVerify: cfg.TLSClientConfig.Insecure,
		},
	}, nil
}

func ToPrometheusConfig(cfg *rest.Config, ref ServiceReference) (*prometheus.Config, error) {
	if err := rest.LoadTLSFiles(cfg); err != nil {
		return nil, err
	}

	certDir, err := os.MkdirTemp(os.TempDir(), "prometheus-*")
	if err != nil {
		return nil, err
	}

	caFile := filepath.Join(certDir, "ca.crt")
	certFile := filepath.Join(certDir, "tls.crt")
	keyFile := filepath.Join(certDir, "tls.key")

	if err := os.WriteFile(caFile, cfg.TLSClientConfig.CAData, 0o644); err != nil {
		return nil, err
	}
	if err := os.WriteFile(certFile, cfg.TLSClientConfig.CertData, 0o644); err != nil {
		return nil, err
	}
	if err := os.WriteFile(keyFile, cfg.TLSClientConfig.KeyData, 0o644); err != nil {
		return nil, err
	}

	return &prometheus.Config{
		Addr: fmt.Sprintf("%s/api/v1/namespaces/%s/services/%s:%s:%d/proxy/", cfg.Host, ref.Namespace, ref.Scheme, ref.Name, ref.Port),
		BasicAuth: prometheus.BasicAuth{
			Username:     cfg.Username,
			Password:     cfg.Password,
			PasswordFile: "",
		},
		BearerToken:     cfg.BearerToken,
		BearerTokenFile: cfg.BearerTokenFile,
		ProxyURL:        "",
		TLSConfig: prom_config.TLSConfig{
			CAFile: caFile,
			// CertFile:           certFile,
			// KeyFile:            keyFile,
			ServerName:         cfg.TLSClientConfig.ServerName,
			InsecureSkipVerify: cfg.TLSClientConfig.Insecure,
		},
	}, nil
}

// ref: https://kubernetes.io/docs/tasks/administer-cluster/access-cluster-services/#manually-constructing-apiserver-proxy-urls
func main() {
	cfg := ctrl.GetConfigOrDie()

	//// k port-forward sts/prometheus-kube-prometheus-stack-prometheus 9090:9090 -n monitoring
	//kc := kubernetes.NewForConfigOrDie(cfg)
	//rw := kc.CoreV1().Services("monitoring").ProxyGet("http", "kube-prometheus-stack-prometheus", "9090", "/api/v1/query", map[string]string{
	//	"query": "up",
	//})
	//data2, err := rw.DoRaw(context.TODO())
	//fmt.Println(string(data2))

	//promConfig, err := ToPrometheusConfig(cfg, ServiceReference{
	//	Scheme:    "http",
	//	Name:      "kube-prometheus-stack-prometheus",
	//	Namespace: "monitoring",
	//	Port:      9090,
	//})
	promConfig, err := ToPrometheusConfigFromServiceAccount(cfg,
		types.NamespacedName{
			Namespace: "default",
			Name:      "trickster",
		},
		ServiceReference{
			Scheme:    "http",
			Name:      "kube-prometheus-stack-prometheus",
			Namespace: "monitoring",
			Port:      9090,
		})
	pc, err := promConfig.NewPrometheusClient()
	if err != nil {
		panic(err)
	}
	pc2 := promv1.NewAPI(pc)

	promCPUQuery := `up`

	res, err := getPromQueryResult(pc2, promCPUQuery)
	if err != nil {
		log.Fatalf("failed to get prometheus cpu query result, reason: %v", err)
	}
	data, _ := json.MarshalIndent(res, "", "  ")
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
