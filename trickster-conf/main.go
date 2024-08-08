package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
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
	"github.com/trickstercache/trickster/v2/cmd/trickster/config/validate"
	bo "github.com/trickstercache/trickster/v2/pkg/backends/options"
	rule "github.com/trickstercache/trickster/v2/pkg/backends/rule/options"
	"github.com/trickstercache/trickster/v2/pkg/cache/negative"
	cache "github.com/trickstercache/trickster/v2/pkg/cache/options"
	fropt "github.com/trickstercache/trickster/v2/pkg/frontend/options"
	lo "github.com/trickstercache/trickster/v2/pkg/observability/logging/options"
	mo "github.com/trickstercache/trickster/v2/pkg/observability/metrics/options"
	tracing "github.com/trickstercache/trickster/v2/pkg/observability/tracing/options"
	rwopts "github.com/trickstercache/trickster/v2/pkg/proxy/request/rewriter/options"
	to "github.com/trickstercache/trickster/v2/pkg/proxy/tls/options"
	"github.com/trickstercache/trickster/v2/pkg/util/yamlx"
	trickstercachev1alpha1 "go.openviz.dev/trickster-config/api/v1alpha1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"
)

func NewClient() (client.Client, error) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = trickstercachev1alpha1.AddToScheme(scheme)

	ctrl.SetLogger(klog.NewKlogr())
	cfg := ctrl.GetConfigOrDie()
	cfg.QPS = 100
	cfg.Burst = 100

	hc, err := rest.HTTPClientFor(cfg)
	if err != nil {
		return nil, err
	}
	mapper, err := apiutil.NewDynamicRESTMapper(cfg, hc)
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

const backendName = "k8s"

func main_gen_cfg() {
	cfg := ctrl.GetConfigOrDie()
	pc, err := prepConfig(cfg, ServiceReference{
		Scheme: "http",
		// Name:      "kube-prometheus-stack-prometheus"
		Name:      "prometheus-kube-prometheus-prometheus",
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
			backendName: {
				Provider:  "prometheus",
				OriginURL: pc.Addr,
				TLS: &to.Options{
					ServeTLS:           false,
					InsecureSkipVerify: false,
					CertificateAuthorityPaths: []string{
						filepath.Join(pwd, "certs", "ca.crt"),
					},
					// ClientCertPath: filepath.Join(pwd, "certs", "tls.crt"),
					// ClientKeyPath:  filepath.Join(pwd, "certs", "tls.key"),
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

	if pc.BasicAuth.Password != "" {
		cfg2.RequestRewriters = map[string]*rwopts.Options{
			backendName: {
				Instructions: rwopts.RewriteList{
					{
						"header",
						"set",
						"Authorization",
						"Basic " + base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", pc.BasicAuth.Username, pc.BasicAuth.Password))),
					},
				},
			},
		}
		cfg2.Backends[backendName].ReqRewriterName = backendName
	} else if pc.BearerToken != "" {
		cfg2.RequestRewriters = map[string]*rwopts.Options{
			backendName: {
				Instructions: rwopts.RewriteList{
					{
						"header",
						"set",
						"Authorization",
						"Bearer " + pc.BearerToken,
					},
				},
			},
		}
		cfg2.Backends[backendName].ReqRewriterName = backendName
	} else {
		cfg2.Backends[backendName].TLS.ClientCertPath = filepath.Join(pwd, "certs", "tls.crt")
		cfg2.Backends[backendName].TLS.ClientKeyPath = filepath.Join(pwd, "certs", "tls.key")
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
	/*
		1-a374b4a1-04e2-4164-b268-4f4799f697ed   36d
		1-be34d9c6-74eb-4bfe-bf22-f57c0065b713   36d
	*/

	pc, err := promapi.NewClient(promapi.Config{
		// Address: "http://127.0.0.1:9090/" + "1-be34d9c6-74eb-4bfe-bf22-f57c0065b713",
		// Address: "https://trickster.appscode.ninja/" + "1-ce471d1a-80e3-4998-b7b5-912dc49afaf0",
		// Address: "http://127.0.0.1:3000/" + "1-ce471d1a-80e3-4998-b7b5-912dc49afaf0",

		// Address: "http://localhost:9090/2.42536768-fac2-4403-aa23-2a3092ffa6c9",
		// Address: "https://172.233.230.15/prometheus/2.42536768-fac2-4403-aa23-2a3092ffa6c9",

		Address: "http://localhost:9090/2.39425934-93cc-4776-be30-1614b52924f4",
		Client:  http.DefaultClient,
	})
	if err != nil {
		panic(err)
	}

	/*
		cfg := ctrl.GetConfigOrDie()
		pcfg, err := prepConfig(cfg, ServiceReference{
			Scheme: "http",
			// Name:      "kube-prometheus-stack-prometheus"
			Name:      "prometheus-kube-prometheus-prometheus",
			Namespace: "monitoring",
			Port:      9090,
		})
		if err != nil {
			panic(err)
		}
		pc, err := pcfg.NewPrometheusClient()
		if err != nil {
			panic(err)
		}
	*/

	/*
		cfg := prometheus.Config{
			Addr:        "https://172.233.230.15/prometheus/1.0e5f7deb-377f-476e-9ee2-280ff24bb7db",
			BearerToken: "2f4c069e7601e9fb7507d5b165f638313eb8530f",
			//TLSConfig: prom_config.TLSConfig{
			//	InsecureSkipVerify: true,
			//},
		}
		pc, err = cfg.NewPrometheusClient()
		if err != nil {
			panic(err)
		}
	*/

	pc2 := promv1.NewAPI(pc)

	promCPUQuery := `up`

	res, err := getPromQueryResult(pc2, promCPUQuery)
	if err != nil {
		klog.Fatalf("failed to get prometheus cpu query result, reason: %v", err)
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
		klog.Infoln("Warning: ", warn)
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

	var caFile, certFile, keyFile string
	caFile = filepath.Join(certDir, "ca.crt")
	if err := os.WriteFile(caFile, cfg.TLSClientConfig.CAData, 0o644); err != nil {
		return nil, err
	}

	if len(cfg.TLSClientConfig.CertData) > 0 {
		certFile = filepath.Join(certDir, "tls.crt")
		if err := os.WriteFile(certFile, cfg.TLSClientConfig.CertData, 0o644); err != nil {
			return nil, err
		}
	}

	if len(cfg.TLSClientConfig.KeyData) > 0 {
		keyFile = filepath.Join(certDir, "tls.key")
		if err := os.WriteFile(keyFile, cfg.TLSClientConfig.KeyData, 0o644); err != nil {
			return nil, err
		}
	}

	return &prometheus.Config{
		Addr: fmt.Sprintf("%s/api/v1/namespaces/%s/services/%s:%s:%d/proxy/", cfg.Host, ref.Namespace, ref.Scheme, ref.Name, ref.Port),
		BasicAuth: prometheus.BasicAuth{
			Username:     cfg.Username,
			Password:     cfg.Password,
			PasswordFile: "",
		},
		BearerToken:     cfg.BearerToken,
		BearerTokenFile: "",
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

func main_gen_crd_config() {
	if err := genCRDConfig(); err != nil {
		panic(err)
	}
}

func genCRDConfig() error {
	kc, err := NewClient()
	if err != nil {
		return err
	}

	r := &TricksterReconciler{
		Client: kc,
		Scheme: kc.Scheme(),
		Fn: func(nc *config.Config) error {
			yml, err := yaml.Marshal(nc)
			if err != nil {
				return err
			}
			md, err := yamlx.GetKeyList(string(yml))
			if err != nil {
				nc.SetDefaults(yamlx.KeyLookup{})
				return err
			}
			err = nc.SetDefaults(md)
			if err != nil {
				return err
				// nc.Main.configFilePath = flags.ConfigPath
				// c.Main.configLastModified = c.CheckFileLastModified()
			}
			// reloadFF(nc, conf, log, wg, caches, args)
			return nil
		},
	}
	_, err = r.Reconcile(context.TODO(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "bb",
			Name:      "bytebuilders",
		},
	})
	return err
}

// const configDir = "/tmp/trickster"
var configDir = func() string {
	d, _ := os.Getwd()
	return d
}()

// TricksterReconciler reconciles a Trickster object
type TricksterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Fn     func(cfg *config.Config) error
}

//+kubebuilder:rbac:groups=trickstercache.org,resources=tricksters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=trickstercache.org,resources=tricksters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=trickstercache.org,resources=tricksters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Trickster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.1/pkg/reconcile
func (r *TricksterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var trickster trickstercachev1alpha1.Trickster
	if err := r.Get(ctx, req.NamespacedName, &trickster); err != nil {
		log.Error(err, "unable to fetch Trickster")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var cfg config.Config
	if trickster.Spec.Main != nil {
		cfg.Main = trickster.Spec.Main
	}
	if trickster.Spec.Nats != nil {
		cfg.Nats = trickster.Spec.Nats
	}
	if trickster.Spec.Secret != nil {
		err := r.writeConfig(ctx, req.Namespace, trickster.Spec.Secret)
		if err != nil {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	}
	if trickster.Spec.Frontend != nil {
		cfg.Frontend = trickster.Spec.Frontend
	}
	if trickster.Spec.Logging != nil {
		cfg.Logging = trickster.Spec.Logging
	}
	if trickster.Spec.Metrics != nil {
		cfg.Metrics = trickster.Spec.Metrics
	}
	if trickster.Spec.NegativeCacheConfigs != nil {
		cfg.NegativeCacheConfigs = trickster.Spec.NegativeCacheConfigs
	}
	if trickster.Spec.ReloadConfig != nil {
		cfg.ReloadConfig = trickster.Spec.ReloadConfig
	}
	{
		var list trickstercachev1alpha1.TricksterBackendList
		sel := labels.Everything()
		if trickster.Spec.BackendSelector != nil {
			var err error
			sel, err = metav1.LabelSelectorAsSelector(trickster.Spec.BackendSelector)
			if err != nil {
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}
		}
		if err := r.List(context.Background(), &list, client.InNamespace(req.Namespace), client.MatchingLabelsSelector{Selector: sel}); err != nil {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		cfg.Backends = make(map[string]*bo.Options, len(list.Items))
		for _, item := range list.Items {
			if item.Spec.Secret != nil {
				err := r.writeConfig(ctx, req.Namespace, item.Spec.Secret)
				if err != nil {
					return ctrl.Result{}, client.IgnoreNotFound(err)
				}
			}
			cfg.Backends[item.Name] = &item.Spec.Options
		}
	}
	{
		var list trickstercachev1alpha1.TricksterCacheList
		sel := labels.Everything()
		if trickster.Spec.CacheSelector != nil {
			var err error
			sel, err = metav1.LabelSelectorAsSelector(trickster.Spec.CacheSelector)
			if err != nil {
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}
		}
		if err := r.List(context.Background(), &list, client.InNamespace(req.Namespace), client.MatchingLabelsSelector{Selector: sel}); err != nil {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		if cfg.Caches == nil {
			cfg.Caches = make(map[string]*cache.Options, len(list.Items))
		}
		for _, item := range list.Items {
			if item.Spec.Secret != nil {
				err := r.writeConfig(ctx, req.Namespace, item.Spec.Secret)
				if err != nil {
					return ctrl.Result{}, client.IgnoreNotFound(err)
				}
			}
			cfg.Caches[item.Name] = &item.Spec.Options
		}
	}
	{
		var list trickstercachev1alpha1.TricksterRequestRewriterList
		sel := labels.Everything()
		if trickster.Spec.RequestRewriterSelector != nil {
			var err error
			sel, err = metav1.LabelSelectorAsSelector(trickster.Spec.RequestRewriterSelector)
			if err != nil {
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}
		}
		if err := r.List(context.Background(), &list, client.InNamespace(req.Namespace), client.MatchingLabelsSelector{Selector: sel}); err != nil {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		if cfg.RequestRewriters == nil {
			cfg.RequestRewriters = make(map[string]*rwopts.Options, len(list.Items))
		}
		for _, item := range list.Items {
			cfg.RequestRewriters[item.Name] = &item.Spec.Options
		}
	}
	{
		var list trickstercachev1alpha1.TricksterRuleList
		sel := labels.Everything()
		if trickster.Spec.RuleSelector != nil {
			var err error
			sel, err = metav1.LabelSelectorAsSelector(trickster.Spec.RuleSelector)
			if err != nil {
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}
		}
		if err := r.List(context.Background(), &list, client.InNamespace(req.Namespace), client.MatchingLabelsSelector{Selector: sel}); err != nil {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		if cfg.Rules == nil {
			cfg.Rules = make(map[string]*rule.Options, len(list.Items))
		}
		for _, item := range list.Items {
			cfg.Rules[item.Name] = &item.Spec.Options
		}
	}
	{
		var list trickstercachev1alpha1.TricksterTracingConfigList
		sel := labels.Everything()
		if trickster.Spec.TracingConfigSelector != nil {
			var err error
			sel, err = metav1.LabelSelectorAsSelector(trickster.Spec.TracingConfigSelector)
			if err != nil {
				return ctrl.Result{}, client.IgnoreNotFound(err)
			}
		}
		if err := r.List(context.Background(), &list, client.InNamespace(req.Namespace), client.MatchingLabelsSelector{Selector: sel}); err != nil {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		if cfg.TracingConfigs == nil {
			cfg.TracingConfigs = make(map[string]*tracing.Options, len(list.Items))
		}
		for _, item := range list.Items {
			if item.Spec.Secret != nil {
				err := r.writeConfig(ctx, req.Namespace, item.Spec.Secret)
				if err != nil {
					return ctrl.Result{}, client.IgnoreNotFound(err)
				}
			}
			cfg.TracingConfigs[item.Name] = &item.Spec.Options
		}
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	fmt.Println("-------------------------")
	fmt.Println(string(data))
	fmt.Println("-------------------------")
	yml := string(data)

	c, err := LoadConfig(yml)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	data2, err := yaml.Marshal(c)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	fmt.Println("-------------------------")
	fmt.Println(string(data2))
	fmt.Println("-------------------------")

	// if err := r.Fn(cfg); err != nil {
	//	return ctrl.Result{}, client.IgnoreNotFound(err)
	//}

	return ctrl.Result{}, nil
}

func LoadConfig(yml string) (*config.Config, error) {
	c := config.NewConfig()
	err := yaml.Unmarshal([]byte(yml), &c)
	if err != nil {
		return nil, err
	}
	md, err := yamlx.GetKeyList(yml)
	if err != nil {
		c.SetDefaults(yamlx.KeyLookup{})
		return nil, err
	}
	err = c.SetDefaults(md)
	//if err == nil {
	//	c.Main.configFilePath = flags.ConfigPath
	//	c.Main.configLastModified = c.CheckFileLastModified()
	//}

	// set the default origin url from the flags
	if d, ok := c.Backends["default"]; ok {
		// If the user has configured their own backends, and one of them is not "default"
		// then Trickster will not use the auto-created default backend
		if d.OriginURL == "" {
			delete(c.Backends, "default")
		}
	}

	if len(c.Backends) == 0 {
		return nil, errors.New("no valid backends configured")
	}

	ncl, err := negative.ConfigLookup(c.NegativeCacheConfigs).Validate()
	if err != nil {
		return nil, err
	}

	err = bo.Lookup(c.Backends).Validate(ncl)
	if err != nil {
		return nil, err
	}

	for _, c := range c.Caches {
		c.Index.FlushInterval = time.Duration(c.Index.FlushIntervalMS) * time.Millisecond
		c.Index.ReapInterval = time.Duration(c.Index.ReapIntervalMS) * time.Millisecond
	}

	err = validate.ValidateConfig(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *TricksterReconciler) writeConfig(ctx context.Context, ns string, sp *core.SecretProjection) error {
	var secret core.Secret
	err := r.Get(ctx, client.ObjectKey{Namespace: ns, Name: sp.Name}, &secret)
	if err != nil {
		return err
	}
	for _, item := range sp.Items {
		path := item.Path
		if !filepath.IsAbs(path) {
			path = filepath.Join(configDir, path)
		}
		err = os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			return err
		}
		err = os.WriteFile(path, secret.Data[item.Key], 0644)
		if err != nil {
			return err
		}
	}
	return nil
}
