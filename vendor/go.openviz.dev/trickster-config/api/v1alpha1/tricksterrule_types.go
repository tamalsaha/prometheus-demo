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
	rule "github.com/trickstercache/trickster/v2/pkg/backends/rule/options"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TricksterRuleSpec defines the desired state of TricksterRule
type TricksterRuleSpec struct {
	rule.Options `json:",inline"`
}

// TricksterRuleStatus defines the observed state of TricksterRule
type TricksterRuleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TricksterRule is the Schema for the tricksterrules API
type TricksterRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TricksterRuleSpec   `json:"spec,omitempty"`
	Status TricksterRuleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TricksterRuleList contains a list of TricksterRule
type TricksterRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TricksterRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TricksterRule{}, &TricksterRuleList{})
}
