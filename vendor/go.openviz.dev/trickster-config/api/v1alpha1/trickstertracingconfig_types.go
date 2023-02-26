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
	tracing "github.com/trickstercache/trickster/v2/pkg/observability/tracing/options"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TricksterTracingConfigSpec defines the desired state of TricksterTracingConfig
type TricksterTracingConfigSpec struct {
	tracing.Options `json:",inline"`
	// secret information about the secret data to project
	// +optional
	Secret *core.SecretProjection `json:"secret,omitempty"`
}

// TricksterTracingConfigStatus defines the observed state of TricksterTracingConfig
type TricksterTracingConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TricksterTracingConfig is the Schema for the trickstertracingconfigs API
type TricksterTracingConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TricksterTracingConfigSpec   `json:"spec,omitempty"`
	Status TricksterTracingConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TricksterTracingConfigList contains a list of TricksterTracingConfig
type TricksterTracingConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TricksterTracingConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TricksterTracingConfig{}, &TricksterTracingConfigList{})
}
