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
	cache "github.com/trickstercache/trickster/v2/pkg/cache/options"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TricksterCacheSpec defines the desired state of TricksterCache
type TricksterCacheSpec struct {
	cache.Options `json:",inline"`
	// secret information about the secret data to project
	// +optional
	Secret *core.SecretProjection `json:"secret,omitempty"`
}

// TricksterCacheStatus defines the observed state of TricksterCache
type TricksterCacheStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TricksterCache is the Schema for the trickstercaches API
type TricksterCache struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TricksterCacheSpec   `json:"spec,omitempty"`
	Status TricksterCacheStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TricksterCacheList contains a list of TricksterCache
type TricksterCacheList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TricksterCache `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TricksterCache{}, &TricksterCacheList{})
}
