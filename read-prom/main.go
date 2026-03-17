package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prom_config "github.com/prometheus/common/config"
	"github.com/tamalsaha/prometheus-demo/prometheus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	cu "kmodules.xyz/client-go/client"
	promclient "kmodules.xyz/monitoring-agent-api/client"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
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
	cc, err := cu.NewUncachedClient(cfg)
	if err != nil {
		return nil, err
	}

	secret, err := cu.GetServiceAccountTokenSecret(cc, sa)
	if err != nil {
		return nil, err
	}
	caData := secret.Data["ca.crt"]
	tokenData := secret.Data["token"]

	certDir, err := os.MkdirTemp(os.TempDir(), "prometheus-*")
	if err != nil {
		return nil, err
	}

	caFile := filepath.Join(certDir, "ca.crt")
	if err := os.WriteFile(caFile, caData, 0o644); err != nil {
		return nil, err
	}

	return &prometheus.Config{
		// Addr:        fmt.Sprintf("%s/api/v1/namespaces/%s/services/%s:%s:%d/proxy/", cfg.Host, ref.Namespace, ref.Scheme, ref.Name, ref.Port),
		Addr:        "https://thanos-querier-openshift-monitoring.apps.pmmswrjj775acdea26.centralindia.aroapp.io",
		ProxyURL:    "",
		BearerToken: string(tokenData),
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
		Addr: "https://rancher01.elogic.cloud/k8s/clusters/c-m-w5q4j76m/api/v1/namespaces/cattle-monitoring-system/services/http:rancher-monitoring-prometheus:9090/proxy/",
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

// https://rancher01.elogic.cloud/k8s/clusters/c-m-w5q4j76m/api/v1/namespaces/cattle-monitoring-system/services/http:rancher-monitoring-prometheus:9090/proxy/

// ref: https://kubernetes.io/docs/tasks/administer-cluster/access-cluster-services/#manually-constructing-apiserver-proxy-urls
func main_() {
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
	//promConfig, err := ToPrometheusConfigFromServiceAccount(cfg,
	//	types.NamespacedName{
	//		Namespace: "monitoring",
	//		Name:      "trickster",
	//	},
	//	ServiceReference{
	//		Scheme:    "http",
	//		Name:      "rancher-monitoring-prometheus",
	//		Namespace: "cattle-monitoring-system",
	//		Port:      9090,
	//	})
	promConfig, err := ToPrometheusConfigFromServiceAccount(cfg,
		types.NamespacedName{
			Namespace: "kubeops",
			Name:      "kube-ui-server",
		},
		ServiceReference{
			Scheme:    "https",
			Namespace: "openshift-monitoring",
			Name:      "thanos-querier",
			Port:      9091,
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

func main() {
	if err := useKubebuilderClient(); err != nil {
		panic(err)
	}
}
func useKubebuilderClient() error {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)

	ctx := ctrl.SetupSignalHandler()

	cfg := ctrl.GetConfigOrDie()
	mgr, err := manager.New(cfg, manager.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: ""},
		HealthProbeBindAddress: "",
		LeaderElection:         false,
		LeaderElectionID:       "5b87adeb.ui-server.kubeops.dev",
	})
	if err != nil {
		return err
	}

	builder, err := promclient.NewBuilder(mgr, nil)
	if err != nil {
		return err
	}
	if err := builder.Setup(); err != nil {
		return err
	}

	mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		pc, err := builder.GetPrometheusClient()
		if err != nil {
			klog.ErrorS(err, "failed to create Prometheus client")
		}
		promCPUQuery := `up`

		res, err := getPromQueryResult(pc, promCPUQuery)
		if err != nil {
			log.Fatalf("failed to get prometheus cpu query result, reason: %v", err)
		}
		data, _ := json.MarshalIndent(res, "", "  ")
		fmt.Println(string(data))
		return nil
	}))

	mgr.Start(ctx)
	select {}
	return nil
}
