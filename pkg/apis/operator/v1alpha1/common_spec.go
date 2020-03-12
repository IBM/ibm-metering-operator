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

// MeteringSpecMongoDB defines the MongoDB configuration in all the Metering specs
type MeteringSpecMongoDB struct {
	Host               string `json:"host"`
	Port               int    `json:"port"`
	UsernameSecret     string `json:"usernameSecret"`
	UsernameKey        string `json:"usernameKey"`
	PasswordSecret     string `json:"passwordSecret"`
	PasswordKey        string `json:"passwordKey"`
	ClusterCertsSecret string `json:"clustercertssecret"`
	ClientCertsSecret  string `json:"clientcertssecret"`
}

// MeteringSpecUI defines the UI configuration in the MeteringUI and MeteringMultiCloudUI specs
type MeteringSpecUI struct {
	IAMnamespace          string `json:"iamNamespace,omitempty"`
	IngressNamespace      string `json:"ingressNamespace,omitempty"`
	CommonHeaderNamespace string `json:"commonHeaderNamespace,omitempty"`
	APIkeySecret          string `json:"apikeySecret,omitempty"`
	PlatformOidcSecret    string `json:"platformOidcSecret,omitempty"`
}

// MeteringStatus defines the observed state of each Metering service
type MeteringStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	// PodNames are the names of the metering pods
	PodNames []string `json:"podNames"`
}
