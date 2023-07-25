package main

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
	"k8s.io/klog/v2/klogr"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	ctrl.SetLogger(klogr.New())
	cfg := ctrl.GetConfigOrDie()
	cfg.QPS = 100
	cfg.Burst = 100

	data, err := GenerateKubeConfiguration(cfg, "default")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}

// https://github.com/kubernetes/client-go/issues/711#issuecomment-730112049
func GenerateKubeConfiguration(cfg *rest.Config, namespace string) ([]byte, error) {
	if err := rest.LoadTLSFiles(cfg); err != nil {
		return nil, err
	}

	clusters := make(map[string]*clientcmdapi.Cluster)
	clusters["default-cluster"] = &clientcmdapi.Cluster{
		Server:                   cfg.Host,
		CertificateAuthorityData: cfg.CAData,
	}

	contexts := make(map[string]*clientcmdapi.Context)
	contexts["default-context"] = &clientcmdapi.Context{
		Cluster:   "default-cluster",
		Namespace: namespace,
		AuthInfo:  "default-user",
	}

	authinfos := make(map[string]*clientcmdapi.AuthInfo)
	authinfos["default-user"] = &clientcmdapi.AuthInfo{
		LocationOfOrigin:      "",
		ClientCertificate:     "",
		ClientCertificateData: cfg.CertData,
		ClientKey:             "",
		ClientKeyData:         cfg.KeyData,
		Token:                 cfg.BearerToken,
		TokenFile:             "",
		Impersonate:           cfg.Impersonate.UserName,
		ImpersonateUID:        cfg.Impersonate.UID,
		ImpersonateGroups:     cfg.Impersonate.Groups,
		ImpersonateUserExtra:  cfg.Impersonate.Extra,
		Username:              cfg.Username,
		Password:              cfg.Password,
		AuthProvider:          cfg.AuthProvider,
		Exec:                  cfg.ExecProvider,
		Extensions:            nil,
	}

	clientConfig := clientcmdapi.Config{
		Kind:           "Config",
		APIVersion:     "v1",
		Clusters:       clusters,
		Contexts:       contexts,
		CurrentContext: "default-context",
		AuthInfos:      authinfos,
	}
	return runtime.Encode(clientcmdlatest.Codec, &clientConfig)
}
