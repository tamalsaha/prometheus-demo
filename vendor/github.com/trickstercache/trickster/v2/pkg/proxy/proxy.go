/*
 * Copyright 2018 The Trickster Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package proxy provides all proxy services for Trickster
package proxy

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	bo "github.com/trickstercache/trickster/v2/pkg/backends/options"
	"github.com/trickstercache/trickster/v2/pkg/proxy/nats"

	ktr "k8s.io/client-go/transport"
	"kubeops.dev/cluster-connector/pkg/shared"
	"kubeops.dev/cluster-connector/pkg/transport"
)

// NewHTTPClient returns an HTTP client configured to the specifications of the
// running Trickster config.
func NewHTTPClient(o *bo.Options) (*http.Client, error) {
	if o == nil {
		return nil, nil
	}

	if o.Transport != nil && o.Transport.Type == bo.NATSTransport {
		tlsConfig := ktr.TLSConfig{
			Insecure: o.TLS.InsecureSkipVerify,
		}
		if o.TLS.ClientCertPath != "" && o.TLS.ClientKeyPath != "" {
			// load client cert
			certPEMBlock, err := os.ReadFile(o.TLS.ClientCertPath)
			if err != nil {
				return nil, err
			}
			tlsConfig.CertData = certPEMBlock

			keyPEMBlock, err := os.ReadFile(o.TLS.ClientKeyPath)
			if err != nil {
				return nil, err
			}
			tlsConfig.KeyData = keyPEMBlock
		}
		if o.TLS.CertificateAuthorityPaths != nil && len(o.TLS.CertificateAuthorityPaths) > 0 {

			// credit snippet to https://forfuncsake.github.io/post/2017/08/trust-extra-ca-cert-in-go-app/
			// Get the SystemCertPool, continue with an empty pool on error
			rootCAs, _ := x509.SystemCertPool()
			if rootCAs == nil {
				rootCAs = x509.NewCertPool()
			}

			var buf bytes.Buffer
			for _, path := range o.TLS.CertificateAuthorityPaths {
				// Read in the cert file
				certs, err := os.ReadFile(path)
				if err != nil {
					return nil, err
				}
				buf.Write(certs)
				buf.WriteRune('\n')
			}

			// Trust the augmented cert pool in our client
			tlsConfig.CAData = buf.Bytes()
		}

		var names shared.SubjectNames
		if o.Transport.CrossAccount {
			names = shared.CrossAccountNames{LinkID: o.Transport.LinkID}
		} else {
			names = shared.SameAccountNames{LinkID: o.Transport.LinkID}
		}

		cfg := &ktr.Config{
			TLS:                tlsConfig,
			DisableCompression: true,
		}
		nc, err := nats.Connection()
		if err != nil {
			return nil, err
		}
		tr, err := transport.New(cfg, nc, names, o.Timeout)
		if err != nil {
			return nil, err
		}
		return &http.Client{
			Timeout: o.Timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: tr,
		}, nil
	}

	var TLSConfig *tls.Config

	if o.TLS != nil {
		TLSConfig = &tls.Config{InsecureSkipVerify: o.TLS.InsecureSkipVerify}

		if o.TLS.ClientCertPath != "" && o.TLS.ClientKeyPath != "" {
			// load client cert
			cert, err := tls.LoadX509KeyPair(o.TLS.ClientCertPath, o.TLS.ClientKeyPath)
			if err != nil {
				return nil, err
			}
			TLSConfig.Certificates = []tls.Certificate{cert}
		}

		if o.TLS.CertificateAuthorityPaths != nil && len(o.TLS.CertificateAuthorityPaths) > 0 {

			// credit snippet to https://forfuncsake.github.io/post/2017/08/trust-extra-ca-cert-in-go-app/
			// Get the SystemCertPool, continue with an empty pool on error
			rootCAs, _ := x509.SystemCertPool()
			if rootCAs == nil {
				rootCAs = x509.NewCertPool()
			}

			for _, path := range o.TLS.CertificateAuthorityPaths {
				// Read in the cert file
				certs, err := os.ReadFile(path)
				if err != nil {
					return nil, err
				}
				// Append our cert to the system pool
				if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
					return nil, fmt.Errorf("unable to append to CA Certs from file %s", path)
				}
			}

			// Trust the augmented cert pool in our client
			TLSConfig.RootCAs = rootCAs
		}
	}

	return &http.Client{
		Timeout: o.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			Dial:                (&net.Dialer{KeepAlive: time.Duration(o.KeepAliveTimeoutMS) * time.Millisecond}).Dial,
			MaxIdleConns:        o.MaxIdleConns,
			MaxIdleConnsPerHost: o.MaxIdleConns,
			TLSClientConfig:     TLSConfig,
		},
	}, nil
}
