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

// MeteringMultiClusterUISpec defines the desired state of MeteringMultiClusterUI
type MeteringMultiClusterUISpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	OperatorVersion  string                                `json:"operatorVersion,omitempty"`
	ImageRegistry    string                                `json:"imageRegistry,omitempty"`
	IAMnamespace     string                                `json:"iamNamespace,omitempty"`
	IngressNamespace string                                `json:"ingressNamespace,omitempty"`
	UI               MeteringMultiClusterUISpecUI          `json:"ui,omitempty"`
	DataManager      MeteringMultiClusterUISpecDataManager `json:"dm,omitempty"`
	MongoDB          MeteringMultiClusterUISpecMongoDB     `json:"mongodb,omitempty"`
	External         MeteringMultiClusterUISpecExternal    `json:"external,omitempty"`
}

// MeteringMultiClusterUISpecUI defines the metering-mcmui configuration in the the MeteringMultiClusterUI spec
type MeteringMultiClusterUISpecUI struct {
	ImageTagPostfix string `json:"imageTagPostfix,omitempty"`
}

// MeteringMultiClusterUISpecDataManager defines the metering-datamanager configuration in the the MeteringMultiClusterUI spec
type MeteringMultiClusterUISpecDataManager struct {
	ImageTagPostfix string `json:"imageTagPostfix,omitempty"`
}

// MeteringMultiClusterUISpecMongoDB defines the MongoDB configuration in the the MeteringMultiClusterUI spec
type MeteringMultiClusterUISpecMongoDB struct {
	Host               string `json:"host,omitempty"`
	Port               string `json:"port,omitempty"`
	UsernameSecret     string `json:"usernameSecret,omitempty"`
	UsernameKey        string `json:"usernameKey,omitempty"`
	PasswordSecret     string `json:"passwordSecret,omitempty"`
	PasswordKey        string `json:"passwordKey,omitempty"`
	ClusterCertsSecret string `json:"clustercertssecret,omitempty"`
	ClientCertsSecret  string `json:"clientcertssecret,omitempty"`
}

// MeteringMultiClusterUISpecExternal defines the external cluster configuration in the the MeteringMultiClusterUI spec
type MeteringMultiClusterUISpecExternal struct {
	ClusterIP   string `json:"clusterIP,omitempty"`
	ClusterPort string `json:"clusterPort,omitempty"`
	ClusterName string `json:"clusterName,omitempty"`
}

// MeteringMultiClusterUIStatus defines the observed state of MeteringMultiClusterUI
type MeteringMultiClusterUIStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Nodes []string `json:"nodes"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MeteringMultiClusterUI is the Schema for the meteringmulticlusteruis API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=meteringmulticlusteruis,scope=Namespaced
type MeteringMultiClusterUI struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MeteringMultiClusterUISpec   `json:"spec,omitempty"`
	Status MeteringMultiClusterUIStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MeteringMultiClusterUIList contains a list of MeteringMultiClusterUI
type MeteringMultiClusterUIList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MeteringMultiClusterUI `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MeteringMultiClusterUI{}, &MeteringMultiClusterUIList{})
}
