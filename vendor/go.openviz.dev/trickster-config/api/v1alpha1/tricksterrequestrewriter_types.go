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
	rwopts "github.com/trickstercache/trickster/v2/pkg/proxy/request/rewriter/options"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TricksterRequestRewriterSpec defines the desired state of TricksterRequestRewriter
type TricksterRequestRewriterSpec struct {
	rwopts.Options `json:",inline"`
}

// TricksterRequestRewriterStatus defines the observed state of TricksterRequestRewriter
type TricksterRequestRewriterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TricksterRequestRewriter is the Schema for the tricksterrequestrewriters API
type TricksterRequestRewriter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TricksterRequestRewriterSpec   `json:"spec,omitempty"`
	Status TricksterRequestRewriterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TricksterRequestRewriterList contains a list of TricksterRequestRewriter
type TricksterRequestRewriterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TricksterRequestRewriter `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TricksterRequestRewriter{}, &TricksterRequestRewriterList{})
}
