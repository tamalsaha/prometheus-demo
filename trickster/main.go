package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	prom_config "github.com/prometheus/common/config"
	"github.com/tamalsaha/prometheus-demo/prometheus"
	"github.com/trickstercache/trickster/v2/cmd/trickster/config"
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

func main() {
	if err := useKubebuilderClient(); err != nil {
		panic(err)
	}
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
	err := os.MkdirAll(certDir, 0755)
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
