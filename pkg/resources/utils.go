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

package resources

import (
	certmgr "github.com/ibm/metering-operator/pkg/apis/certmanager/v1alpha1"

	"os"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type CertificateData struct {
	Name      string
	Secret    string
	Common    string
	App       string
	Component string
}

type IngressData struct {
	Name        string
	Path        string
	Service     string
	Port        int32
	Annotations map[string]string
}

const MeteringComponentName = "meteringsvc"
const MeteringReleaseName = "metering"
const DmDeploymentName = "metering-dm"
const ReaderDaemonSetName = "metering-reader"
const ServerServiceName = "metering-server"
const UIDeploymentName = "metering-ui"
const McmDeploymentName = "metering-mcmui"
const apiIngressPort int32 = 4000

//CS??? move UI and MCM names here

var DefaultMode int32 = 420

var APICertificateData = CertificateData{
	Name:      APICertName,
	Secret:    APICertZecretName,
	Common:    APICertCommonName,
	App:       ReaderDaemonSetName,
	Component: ReaderDaemonSetName,
}
var ReceiverCertificateData = CertificateData{
	Name:      ReceiverCertName,
	Secret:    ReceiverCertZecretName,
	Common:    ReceiverCertCommonName,
	App:       DmDeploymentName,
	Component: ReceiverCertCommonName,
}

var CommonIngressAnnotations = map[string]string{
	"app.kubernetes.io/managed-by": "operator",
	"kubernetes.io/ingress.class":  "ibm-icp-management",
}
var apiCheckIngressAnnotations = map[string]string{
	"icp.management.ibm.com/location-modifier": "=",
	"icp.management.ibm.com/upstream-uri":      "/api/v1",
}
var apiRBACIngressAnnotations = map[string]string{
	"icp.management.ibm.com/authz-type":     "rbac",
	"icp.management.ibm.com/rewrite-target": "/api",
}
var apiSwaggerIngressAnnotations = map[string]string{
	"icp.management.ibm.com/location-modifier": "=",
	"icp.management.ibm.com/upstream-uri":      "/api/swagger",
}
var uiIngressAnnotations = map[string]string{
	"icp.management.ibm.com/auth-type":      "id-token",
	"icp.management.ibm.com/rewrite-target": "/",
}
var mcmIngressAnnotations = map[string]string{
	"icp.management.ibm.com/auth-type": "id-token",
}

var APIcheckIngressData = IngressData{
	Name:        "metering-api-check",
	Path:        "/meteringapi/api/v1",
	Service:     ServerServiceName,
	Port:        apiIngressPort,
	Annotations: apiCheckIngressAnnotations,
}
var APIrbacIngressData = IngressData{
	Name:        "metering-api-rbac",
	Path:        "/meteringapi/api/",
	Service:     ServerServiceName,
	Port:        apiIngressPort,
	Annotations: apiRBACIngressAnnotations,
}
var APIswaggerIngressData = IngressData{
	Name:        "metering-api-swagger",
	Path:        "/meteringapi/api/swagger",
	Service:     ServerServiceName,
	Port:        apiIngressPort,
	Annotations: apiSwaggerIngressAnnotations,
}
var UIIngressData = IngressData{
	Name:        "metering-ui",
	Path:        "/metering/",
	Service:     "metering-ui",
	Port:        3130,
	Annotations: uiIngressAnnotations,
}
var McmIngressData = IngressData{
	Name:        "metering-mcmui",
	Path:        "/metering-mcm",
	Service:     "metering-mcmui",
	Port:        3001,
	Annotations: mcmIngressAnnotations,
}

var log = logf.Log.WithName("resource_utils")

// BuildCertificate returns a Certificate object.
// Call controllerutil.SetControllerReference to set the owner and controller
// for the Certificate object created by this function.
func BuildCertificate(namespace string, certData CertificateData) *certmgr.Certificate {
	reqLogger := log.WithValues("func", "BuildCertificate", "Certificate.Name", certData.Name)
	reqLogger.Info("CS??? Entry")
	metaLabels := labelsForCertificateMeta(certData.App, certData.Component)

	certificate := &certmgr.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      certData.Name,
			Labels:    metaLabels,
			Namespace: namespace,
		},
		Spec: certmgr.CertificateSpec{
			CommonName: certData.Common,
			SecretName: certData.Secret,
			IsCA:       false,
			DNSNames: []string{
				certData.Common,
				certData.Common + "." + namespace + ".svc.cluster.local",
			},
			Organization: []string{"IBM"},
			IssuerRef: certmgr.ObjectReference{
				Name: "icp-ca-issuer",
				Kind: certmgr.ClusterIssuerKind,
			},
		},
	}
	return certificate
}

// BuildIngress returns an Ingress object.
// Call controllerutil.SetControllerReference to set the owner and controller
// for the Ingress object created by this function.
func BuildIngress(namespace string, ingressData IngressData) *netv1.Ingress {
	reqLogger := log.WithValues("func", "buildIngress", "Ingress.Name", ingressData.Name)
	reqLogger.Info("CS??? Entry")
	metaLabels := labelsForIngressMeta(ingressData.Name)
	newAnnotations := ingressData.Annotations
	for key, value := range CommonIngressAnnotations {
		newAnnotations[key] = value
	}

	ingress := &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        ingressData.Name,
			Annotations: newAnnotations,
			Labels:      metaLabels,
			Namespace:   namespace,
		},
		Spec: netv1.IngressSpec{
			Rules: []netv1.IngressRule{
				{
					IngressRuleValue: netv1.IngressRuleValue{
						HTTP: &netv1.HTTPIngressRuleValue{
							Paths: []netv1.HTTPIngressPath{
								{
									Path: ingressData.Path,
									Backend: netv1.IngressBackend{
										ServiceName: ingressData.Service,
										ServicePort: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: ingressData.Port,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return ingress
}

func BuildMongoDBEnvVars(host string, port string, usernameSecret string, usernameKey string,
	passwordSecret string, passwordKey string) []corev1.EnvVar {
	mongoDBEnvVars := []corev1.EnvVar{
		{
			Name:  "HC_MONGO_HOST",
			Value: host,
		},
		{
			Name:  "HC_MONGO_PORT",
			Value: port,
		},
		{
			Name: "HC_MONGO_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: usernameSecret,
					},
					Key:      usernameKey,
					Optional: &TrueVar,
				},
			},
		},
		{
			Name: "HC_MONGO_PASS",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: passwordSecret,
					},
					Key:      passwordKey,
					Optional: &TrueVar,
				},
			},
		},
		{
			Name:  "HC_MONGO_ISSSL",
			Value: "true",
		},
		{
			Name:  "HC_MONGO_SSL_CA",
			Value: "/certs/mongodb-ca/tls.crt",
		},
		{
			Name:  "HC_MONGO_SSL_CERT",
			Value: "/certs/mongodb-client/tls.crt",
		},
		{
			Name:  "HC_MONGO_SSL_KEY",
			Value: "/certs/mongodb-client/tls.key",
		},
	}
	return mongoDBEnvVars
}

func BuildCommonClusterEnvVars(instanceNamespace, instanceIAMnamespace, clusterName string) []corev1.EnvVar {
	reqLogger := log.WithValues("func", "BuildCommonClusterEnvVars")
	reqLogger.Info("CS??? Entry")

	var iamNamespace string
	if instanceIAMnamespace != "" {
		reqLogger.Info("CS??? IAMnamespace=" + instanceIAMnamespace)
		iamNamespace = instanceIAMnamespace
	} else {
		reqLogger.Info("CS??? IAMnamespace is blank, use instance=" + instanceNamespace)
		iamNamespace = instanceNamespace
	}
	clusterEnvVars := []corev1.EnvVar{
		{
			Name:  "CLUSTER_NAME",
			Value: clusterName,
		},
		{
			Name:  "IAM_NAMESPACE",
			Value: iamNamespace,
		},
	}
	return clusterEnvVars
}

// set isMcmUI to true when building env vars for metering-mcmui.
// set isMcmUI to false when building env vars for any other component.
func BuildUIClusterEnvVars(instanceNamespace, instanceIAMnamespace, instanceIngressNamespace,
	clusterName, clusterIP, clusterPort string, isMcmUI bool) []corev1.EnvVar {

	reqLogger := log.WithValues("func", "BuildUIClusterEnvVars")
	reqLogger.Info("CS??? Entry")

	var iamNamespace string
	if instanceIAMnamespace != "" {
		reqLogger.Info("CS??? IAMnamespace=" + instanceIAMnamespace)
		iamNamespace = instanceIAMnamespace
	} else {
		reqLogger.Info("CS??? IAMnamespace is blank, use instance=" + instanceNamespace)
		iamNamespace = instanceNamespace
	}
	var ingressNamespace string
	if instanceIngressNamespace != "" {
		reqLogger.Info("CS??? IngressNamespace=" + instanceIngressNamespace)
		ingressNamespace = instanceIngressNamespace
	} else {
		reqLogger.Info("CS??? IAMnamespace is blank, use instance=" + instanceNamespace)
		ingressNamespace = instanceNamespace
	}

	clusterEnvVars := BuildCommonClusterEnvVars(instanceNamespace, instanceIAMnamespace, clusterName)

	// CS??? https://icp-management-ingress:443
	// CS??? https://icp-management-ingress.NAMESPACE.svc.cluster.local:443
	cfcRouterURL := "https://icp-management-ingress." + ingressNamespace + ".svc.cluster.local:443"
	envVar := corev1.EnvVar{
		Name:  "cfcRouterUrl",
		Value: cfcRouterURL,
	}
	clusterEnvVars = append(clusterEnvVars, envVar)

	if clusterIP != "" {
		envVar := corev1.EnvVar{
			Name:  "ICP_EXTERNAL_IP",
			Value: clusterIP,
		}
		clusterEnvVars = append(clusterEnvVars, envVar)
	}
	if clusterPort != "" {
		envVar := corev1.EnvVar{
			Name:  "ICP_EXTERNAL_PORT",
			Value: clusterPort,
		}
		clusterEnvVars = append(clusterEnvVars, envVar)
	}
	if isMcmUI {
		envVar := corev1.EnvVar{
			Name:  "PLATFORM_IDENTITY_PROVIDER_URL",
			Value: cfcRouterURL + "/idprovider",
		}
		clusterEnvVars = append(clusterEnvVars, envVar)
	} else {
		envVar := corev1.EnvVar{
			Name:  "PLATFORM_IDENTITY_PROVIDER_URL",
			Value: "https://platform-identity-provider." + iamNamespace + ".svc.cluster.local:4300",
		}
		clusterEnvVars = append(clusterEnvVars, envVar)
	}
	return clusterEnvVars
}

// set loglevelType to "log4js" when building volumes for metering-mcmui.
// set loglevelType to "loglevel" when building volumes for metering-mcmui.
func BuildCommonVolumes(clusterSecret, clientSecret, usernameSecret,
	passwordSecret, loglevelPrefix, loglevelType string) []corev1.Volume {

	// example for metering-ui
	//   Name: loglevel
	//     Key: metering-ui-loglevel.json
	//     Path: loglevel.json
	// example for metering-mcmui
	//   Name: log4js
	//     Key: metering-mcmui-log4js.json
	//     Path: log4js.json
	loglevelKey := loglevelPrefix + "-" + loglevelType + ".json"
	loglevelPath := loglevelType + ".json"

	// CS??? removed icp-serviceid-apikey-secret
	commonVolumes := []corev1.Volume{
		{
			Name: "mongodb-ca-cert",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  clusterSecret,
					DefaultMode: &DefaultMode,
					Optional:    &TrueVar,
				},
			},
		},
		{
			Name: "mongodb-client-cert",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  clientSecret,
					DefaultMode: &DefaultMode,
					Optional:    &TrueVar,
				},
			},
		},
		{
			Name: "muser-icp-mongodb-admin",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  usernameSecret,
					DefaultMode: &DefaultMode,
					Optional:    &TrueVar,
				},
			},
		},
		{
			Name: "mpass-icp-mongodb-admin",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  passwordSecret,
					DefaultMode: &DefaultMode,
					Optional:    &TrueVar,
				},
			},
		},
		{
			Name: loglevelType,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "metering-logging-configuration",
					},
					Items: []corev1.KeyToPath{
						{
							Key:  loglevelKey,
							Path: loglevelPath,
						},
					},
					DefaultMode: &DefaultMode,
					Optional:    &TrueVar,
				},
			},
		},
	}
	return commonVolumes
}

// returns the labels associated with the resource being created
func LabelsForMetadata(deploymentName string) map[string]string {
	return map[string]string{"app.kubernetes.io/name": deploymentName, "app.kubernetes.io/component": MeteringComponentName,
		"app.kubernetes.io/managed-by": "operator", "app.kubernetes.io/instance": MeteringReleaseName, "release": MeteringReleaseName}
}

// returns the labels for selecting the resources belonging to the given metering CR name
func LabelsForSelector(deploymentName string, crType string, crName string) map[string]string {
	return map[string]string{"app": deploymentName, "component": MeteringComponentName, crType: crName}
}

// returns the labels associated with the Pod being created
func LabelsForPodMetadata(deploymentName string, crType string, crName string) map[string]string {
	podLabels := LabelsForMetadata(deploymentName)
	selectorLabels := LabelsForSelector(deploymentName, crType, crName)
	for key, value := range selectorLabels {
		podLabels[key] = value
	}
	return podLabels
}

// returns the labels associated with the Ingress being created
func labelsForIngressMeta(ingressName string) map[string]string {
	return map[string]string{"app.kubernetes.io/name": ingressName, "app.kubernetes.io/instance": MeteringReleaseName,
		"app.kubernetes.io/managed-by": "operator", "release": MeteringReleaseName}
}

func labelsForCertificateMeta(appName, componentName string) map[string]string {
	return map[string]string{"app": appName, "component": componentName, "release": MeteringReleaseName}
}

// GetPodNames returns the pod names of the array of pods passed in
func GetPodNames(pods []corev1.Pod) []string {
	reqLogger := log.WithValues("func", "getPodNames")
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
		reqLogger.Info("CS??? pod name=" + pod.Name)
	}
	return podNames
}

// returns the service account name or default if it is not set in the environment
func GetServiceAccountName() string {

	sa := "default"

	envSa := os.Getenv("SA_NAME")
	if len(envSa) > 0 {
		sa = envSa
	}
	return sa
}
