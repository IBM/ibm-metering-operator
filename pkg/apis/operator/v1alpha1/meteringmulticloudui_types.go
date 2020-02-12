//
// Copyright 2020 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MeteringMultiCloudUISpec defines the desired state of MeteringMultiCloudUI
type MeteringMultiCloudUISpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Version          string                             `json:"version,omitempty"`
	ImageRegistry    string                             `json:"imageRegistry,omitempty"`
	ImageTagPostfix  string                             `json:"imageTagPostfix,omitempty"`
	IAMnamespace     string                             `json:"iamNamespace,omitempty"`
	IngressNamespace string                             `json:"ingressNamespace,omitempty"`
	MongoDB          MeteringMultiCloudUISpecMongoDB  `json:"mongodb,omitempty"`
	External         MeteringMultiCloudUISpecExternal `json:"external,omitempty"`
}

// MeteringMultiCloudUISpecMongoDB defines the MongoDB configuration in the the MeteringMultiCloudUI spec
type MeteringMultiCloudUISpecMongoDB struct {
	Host               string `json:"host,omitempty"`
	Port               string `json:"port,omitempty"`
	UsernameSecret     string `json:"usernameSecret,omitempty"`
	UsernameKey        string `json:"usernameKey,omitempty"`
	PasswordSecret     string `json:"passwordSecret,omitempty"`
	PasswordKey        string `json:"passwordKey,omitempty"`
	ClusterCertsSecret string `json:"clustercertssecret,omitempty"`
	ClientCertsSecret  string `json:"clientcertssecret,omitempty"`
}

// MeteringMultiCloudUISpecExternal defines the external cluster configuration in the the MeteringMultiCloudUI spec
type MeteringMultiCloudUISpecExternal struct {
	ClusterIP   string `json:"clusterIP,omitempty"`
	ClusterPort string `json:"clusterPort,omitempty"`
	ClusterName string `json:"clusterName,omitempty"`
}

// MeteringMultiCloudUIStatus defines the observed state of MeteringMultiCloudUI
type MeteringMultiCloudUIStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Nodes []string `json:"nodes"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MeteringMultiCloudUI is the Schema for the meteringmulticlouduis API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=meteringmulticlouduis,scope=Namespaced
type MeteringMultiCloudUI struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MeteringMultiCloudUISpec   `json:"spec,omitempty"`
	Status MeteringMultiCloudUIStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MeteringMultiCloudUIList contains a list of MeteringMultiCloudUI
type MeteringMultiCloudUIList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MeteringMultiCloudUI `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MeteringMultiCloudUI{}, &MeteringMultiCloudUIList{})
}
