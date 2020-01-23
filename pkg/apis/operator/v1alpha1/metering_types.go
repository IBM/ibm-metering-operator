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

// MeteringSpec defines the desired state of Metering
type MeteringSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	OperatorVersion string                  `json:"operatorVersion,omitempty"`
	DataManager     MeteringSpecDataManager `json:"dm,omitempty"`
	Reader          MeteringSpecReader      `json:"reader,omitempty"`
	MongoDB         MeteringSpecMongoDB     `json:"mongodb,omitempty"`
	External        MeteringSpecExternal    `json:"external,omitempty"`
}

// MeteringSpecDataManager defines the metering-datamanager configuration in the the metering spec
type MeteringSpecDataManager struct {
	ImageRegistry   string `json:"imageRegistry,omitempty"`
	ImageTagPostfix string `json:"imageTagPostfix,omitempty"`
}

// MeteringSpecReader defines the metering-reader configuration in the the metering spec
type MeteringSpecReader struct {
	ImageRegistry   string `json:"imageRegistry,omitempty"`
	ImageTagPostfix string `json:"imageTagPostfix,omitempty"`
}

// MeteringSpecMongoDB defines the MongoDB configuration in the the metering spec
type MeteringSpecMongoDB struct {
	Host               string `json:"host,omitempty"`
	Port               string `json:"port,omitempty"`
	UsernameSecret     string `json:"usernameSecret,omitempty"`
	UsernameKey        string `json:"usernameKey,omitempty"`
	PasswordSecret     string `json:"passwordSecret,omitempty"`
	PasswordKey        string `json:"passwordKey,omitempty"`
	ClusterCertsSecret string `json:"clustercertssecret,omitempty"`
	ClientCertsSecret  string `json:"clientcertssecret,omitempty"`
}

// MeteringSpecExternal defines the external cluster configuration in the the metering spec
type MeteringSpecExternal struct {
	ClusterIP    string `json:"clusterIP,omitempty"`
	ClusterPort  string `json:"clusterPort,omitempty"`
	ClusterName  string `json:"clusterName,omitempty"`
	CfcRouterURL string `json:"cfcRouterUrl,omitempty"`
}

// MeteringStatus defines the observed state of Metering
type MeteringStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	// Nodes are the names of the metering pods
	Nodes []string `json:"nodes"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Metering is the Schema for the meterings API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=meterings,scope=Namespaced
type Metering struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MeteringSpec   `json:"spec,omitempty"`
	Status MeteringStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MeteringList contains a list of Metering
type MeteringList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Metering `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Metering{}, &MeteringList{})
}
