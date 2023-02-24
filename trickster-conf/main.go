package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	promapi "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prom_config "github.com/prometheus/common/config"
	"github.com/tamalsaha/prometheus-demo/prometheus"
	"github.com/trickstercache/trickster/v2/cmd/trickster/config"
	bo "github.com/trickstercache/trickster/v2/pkg/backends/options"
	fropt "github.com/trickstercache/trickster/v2/pkg/frontend/options"
	lo "github.com/trickstercache/trickster/v2/pkg/observability/logging/options"
	mo "github.com/trickstercache/trickster/v2/pkg/observability/metrics/options"
	to "github.com/trickstercache/trickster/v2/pkg/proxy/tls/options"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2/klogr"
	kmapi "kmodules.xyz/client-go/api/v1"
	au "kmodules.xyz/client-go/client/apiutil"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/yaml"
)

func NewClient() (client.Client, error) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)

	ctrl.SetLogger(klogr.New())
	cfg := ctrl.GetConfigOrDie()
	cfg.QPS = 100
	cfg.Burst = 100

	mapper, err := apiutil.NewDynamicRESTMapper(cfg)
	if err != nil {
		return nil, err
	}

	return client.New(cfg, client.Options{
		Scheme: scheme,
		Mapper: mapper,
		//Opts: client.WarningHandlerOptions{
		//	SuppressWarnings:   false,
		//	AllowDuplicateLogs: false,
		//},
	})
}

func main_gen_config() {
	cfg := ctrl.GetConfigOrDie()
	pc, err := prepConfig(cfg, ServiceReference{
		Scheme:    "http",
		Name:      "kube-prometheus-stack-prometheus",
		Namespace: "monitoring",
		Port:      9090,
	})
	if err != nil {
		panic(err)
	}
	//data, err := yaml.Marshal(pc)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(string(data))

	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	cfg2 := config.Config{
		Frontend: &fropt.Options{
			ListenPort: 9090,
		},
		Backends: map[string]*bo.Options{
			"default": {
				Provider:  "prometheus",
				OriginURL: pc.Addr,
				TLS: &to.Options{
					ServeTLS:           false,
					InsecureSkipVerify: false,
					CertificateAuthorityPaths: []string{
						filepath.Join(pwd, "certs", "ca.crt"),
					},
					ClientCertPath: filepath.Join(pwd, "certs", "tls.crt"),
					ClientKeyPath:  filepath.Join(pwd, "certs", "tls.key"),
				},
			},
		},
		Metrics: &mo.Options{
			ListenPort: 8481,
		},
		Logging: &lo.Options{
			LogLevel: "info",
		},
	}

	data, err := yaml.Marshal(cfg2)
	if err != nil {
		panic(err)
	}

	// /Users/tamal/go/src/github.com/tamalsaha/prometheus-demo/trickster-conf
	err = os.WriteFile("trickster-conf/config.yaml", data, 0o644)
	if err != nil {
		panic(err)
	}
}

func main() {
	pc, err := promapi.NewClient(promapi.Config{
		// Address: "http://127.0.0.1:9090/1-a374b4a1-04e2-4164-b268-4f4799f697ed",
		Address: "http://127.0.0.1:9090",
		Client:  http.DefaultClient,
	})
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

func useKubebuilderClient() error {
	fmt.Println("Using kubebuilder client")
	kc, err := NewClient()
	if err != nil {
		return err
	}

	var list core.PodList
	err = kc.List(context.TODO(), &list)
	if err != nil {
		return err
	}
	for _, db := range list.Items {
		fmt.Println(client.ObjectKeyFromObject(&db))
	}
	images := map[string]kmapi.ImageInfo{}
	for _, pod := range list.Items {
		images, err = au.CollectImageInfo(kc, &pod, images)
		if err != nil {
			return err
		}
	}

	data, err := yaml.Marshal(images)
	if err != nil {
		return err
	}
	fmt.Println(string(data))

	return nil
}

var __tmpDir = func() string {
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

func prepConfig(cfg *rest.Config, ref ServiceReference) (*prometheus.Config, error) {
	if err := rest.LoadTLSFiles(cfg); err != nil {
		return nil, err
	}

	certDir := "certs"
	err := os.MkdirAll(certDir, 0o755)
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
			CAFile:             caFile,
			CertFile:           certFile,
			KeyFile:            keyFile,
			ServerName:         cfg.TLSClientConfig.ServerName,
			InsecureSkipVerify: cfg.TLSClientConfig.Insecure,
		},
	}, nil
}

func TC() (*config.Config, error) {
	tc := config.Config{}
	return &tc, nil
}
