/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"github.com/trickstercache/trickster/v2/cmd/trickster/config"
	reload "github.com/trickstercache/trickster/v2/cmd/trickster/config/reload/options"
	"github.com/trickstercache/trickster/v2/pkg/cache/negative"
	fropt "github.com/trickstercache/trickster/v2/pkg/frontend/options"
	lo "github.com/trickstercache/trickster/v2/pkg/observability/logging/options"
	mo "github.com/trickstercache/trickster/v2/pkg/observability/metrics/options"
	no "github.com/trickstercache/trickster/v2/pkg/proxy/nats/options"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TricksterSpec defines the desired state of Trickster
type TricksterSpec struct {
	// Main is the primary MainConfig section
	Main *config.MainConfig `json:"main,omitempty"`
	// Nats is provides for transport via NATS.io
	Nats *no.Options `json:"nats,omitempty"`
	// secret information about the secret data to project
	// +optional
	Secret *core.SecretProjection `json:"secret,omitempty"`
	// Backends is a map of BackendOptionss
	// Backends map[string]*bo.Options `json:"backends,omitempty"`
	BackendSelector *metav1.LabelSelector `json:"backend_selector,omitempty"`
	// Caches is a map of CacheConfigs
	// Caches map[string]*cache.Options `json:"caches,omitempty"`
	CacheSelector *metav1.LabelSelector `json:"cache_selector,omitempty"`
	// ProxyServer is provides configurations about the Proxy Front End
	Frontend *fropt.Options `json:"frontend,omitempty"`
	// Logging provides configurations that affect logging behavior
	Logging *lo.Options `json:"logging,omitempty"`
	// Metrics provides configurations for collecting Metrics about the application
	Metrics *mo.Options `json:"metrics,omitempty"`
	// TracingConfigs provides the distributed tracing configuration
	// TracingConfigs map[string]*tracing.Options `json:"tracing,omitempty"`
	TracingConfigSelector *metav1.LabelSelector `json:"tracing_config_selector,omitempty"`
	// NegativeCacheConfigs is a map of NegativeCacheConfigs
	NegativeCacheConfigs map[string]negative.Config `json:"negative_caches,omitempty"`
	// Rules is a map of the Rules
	// Rules map[string]*rule.Options `json:"rules,omitempty"`
	RuleSelector *metav1.LabelSelector `json:"rule_selector,omitempty"`
	// RequestRewriters is a map of the Rewriters
	// RequestRewriters map[string]*rwopts.Options `json:"request_rewriters,omitempty"`
	RequestRewriterSelector *metav1.LabelSelector `json:"request_rewriter_selector,omitempty"`
	// ReloadConfig provides configurations for in-process config reloading
	ReloadConfig *reload.Options `json:"reloading,omitempty"`

	//// Resources holds runtime resources uses by the Config
	//Resources *config.Resources `json:"-"`
	//
	//CompiledRewriters map[string]rewriter.RewriteInstructions `json:"-"`
	//activeCaches      map[string]interface{}
	//providedOriginURL string
	//providedProvider  string
	//
	//LoaderWarnings []string `json:"-"`
}

// TricksterStatus defines the observed state of Trickster
type TricksterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Trickster is the Schema for the tricksters API
type Trickster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TricksterSpec   `json:"spec,omitempty"`
	Status TricksterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TricksterList contains a list of Trickster
type TricksterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Trickster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Trickster{}, &TricksterList{})
}
