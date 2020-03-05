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
	"strconv"

	operatorv1alpha1 "github.com/ibm/ibm-metering-operator/pkg/apis/operator/v1alpha1"
	certmgr "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"

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
const MeteringProductName = "IBM Cloud Platform Common Services"
const MeteringProductID = "1234-567"
const MeteringProductVersion = "3.5.0"
const DmDeploymentName = "metering-dm"
const DmServiceName = "metering-dm"
const ReaderDaemonSetName = "metering-reader"
const ReaderServiceName = "metering-server"
const UIDeploymentName = "metering-ui"
const UIServiceName = "metering-ui"
const McmDeploymentName = "metering-mcmui"
const McmServiceName = "metering-mcmui"
const SenderDeploymentName = "metering-sender"
const ReceiverServiceName = "metering-receiver"
const apiIngressPort int32 = 4000

var DefaultMode int32 = 420

var APICertificateData = CertificateData{
	Name:      APICertName,
	Secret:    APICertSecretName,
	Common:    APICertCommonName,
	App:       ReaderDaemonSetName,
	Component: ReaderDaemonSetName,
}
var ReceiverCertificateData = CertificateData{
	Name:      ReceiverCertName,
	Secret:    ReceiverCertSecretName,
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
	Service:     ReaderServiceName,
	Port:        apiIngressPort,
	Annotations: apiCheckIngressAnnotations,
}
var APIrbacIngressData = IngressData{
	Name:        "metering-api-rbac",
	Path:        "/meteringapi/api/",
	Service:     ReaderServiceName,
	Port:        apiIngressPort,
	Annotations: apiRBACIngressAnnotations,
}
var APIswaggerIngressData = IngressData{
	Name:        "metering-api-swagger",
	Path:        "/meteringapi/api/swagger",
	Service:     ReaderServiceName,
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
func BuildCertificate(instanceNamespace, instanceClusterIssuer string, certData CertificateData) *certmgr.Certificate {
	reqLogger := log.WithValues("func", "BuildCertificate")

	metaLabels := labelsForCertificateMeta(certData.App, certData.Component)
	var clusterIssuer string
	if instanceClusterIssuer != "" {
		reqLogger.Info("clusterIssuer=" + instanceClusterIssuer)
		clusterIssuer = instanceClusterIssuer
	} else {
		reqLogger.Info("clusterIssuer is blank, default=" + DefaultClusterIssuer)
		clusterIssuer = DefaultClusterIssuer
	}

	certificate := &certmgr.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      certData.Name,
			Labels:    metaLabels,
			Namespace: instanceNamespace,
		},
		Spec: certmgr.CertificateSpec{
			CommonName: certData.Common,
			SecretName: certData.Secret,
			IsCA:       false,
			DNSNames: []string{
				certData.Common,
				certData.Common + "." + instanceNamespace,
				certData.Common + "." + instanceNamespace + ".svc.cluster.local",
			},
			Organization: []string{"IBM"},
			IssuerRef: certmgr.ObjectReference{
				Name: clusterIssuer,
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

func BuildMongoDBEnvVars(mongoDB operatorv1alpha1.MeteringSpecMongoDB) []corev1.EnvVar {
	mongoDBEnvVars := []corev1.EnvVar{
		{
			Name:  "HC_MONGO_HOST",
			Value: mongoDB.Host,
		},
		{
			Name:  "HC_MONGO_PORT",
			Value: strconv.Itoa(mongoDB.Port),
		},
		{
			Name: "HC_MONGO_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: mongoDB.UsernameSecret,
					},
					Key:      mongoDB.UsernameKey,
					Optional: &TrueVar,
				},
			},
		},
		{
			Name: "HC_MONGO_PASS",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: mongoDB.PasswordSecret,
					},
					Key:      mongoDB.PasswordKey,
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

func BuildCommonClusterEnvVars(instanceNamespace, instanceIAMnamespace string) []corev1.EnvVar {
	reqLogger := log.WithValues("func", "BuildCommonClusterEnvVars")

	var iamNamespace string
	if instanceIAMnamespace != "" {
		reqLogger.Info("IAMnamespace=" + instanceIAMnamespace)
		iamNamespace = instanceIAMnamespace
	} else {
		reqLogger.Info("IAMnamespace is blank, use instance=" + instanceNamespace)
		iamNamespace = instanceNamespace
	}

	clusterEnvVars := []corev1.EnvVar{
		{
			Name:  "IAM_NAMESPACE",
			Value: iamNamespace,
		},
	}
	return clusterEnvVars
}

// set isMcmUI to true when building env vars for metering-mcmui.
// set isMcmUI to false when building env vars for any other component.
func BuildUIClusterEnvVars(instanceNamespace, instanceClusterName string,
	instanceUI operatorv1alpha1.MeteringSpecUI, isMcmUI bool) []corev1.EnvVar {

	reqLogger := log.WithValues("func", "BuildUIClusterEnvVars")

	var iamNamespace string
	if instanceUI.IAMnamespace != "" {
		reqLogger.Info("IAMnamespace=" + instanceUI.IAMnamespace)
		iamNamespace = instanceUI.IAMnamespace
	} else {
		reqLogger.Info("IAMnamespace is blank, use instance=" + instanceNamespace)
		iamNamespace = instanceNamespace
	}
	var ingressNamespace string
	if instanceUI.IngressNamespace != "" {
		reqLogger.Info("IngressNamespace=" + instanceUI.IngressNamespace)
		ingressNamespace = instanceUI.IngressNamespace
	} else {
		reqLogger.Info("IngressNamespace is blank, use instance=" + instanceNamespace)
		ingressNamespace = instanceNamespace
	}
	var headerNamespace string
	if instanceUI.CommonHeaderNamespace != "" {
		reqLogger.Info("HeaderNamespace=" + instanceUI.CommonHeaderNamespace)
		headerNamespace = instanceUI.CommonHeaderNamespace
	} else {
		reqLogger.Info("HeaderNamespace is blank, use instance=" + instanceNamespace)
		headerNamespace = instanceNamespace
	}
	var apiKeySecretName string
	if instanceUI.APIkeySecret != "" {
		reqLogger.Info("apiKeySecretName=" + instanceUI.APIkeySecret)
		apiKeySecretName = instanceUI.APIkeySecret
	} else {
		reqLogger.Info("apiKeySecretName is blank, default=" + DefaultAPIKeySecretName)
		apiKeySecretName = DefaultAPIKeySecretName
	}
	var platformOidcSecretName string
	if instanceUI.PlatformOidcSecret != "" {
		reqLogger.Info("platformOidcSecretName=" + instanceUI.PlatformOidcSecret)
		platformOidcSecretName = instanceUI.PlatformOidcSecret
	} else {
		reqLogger.Info("platformOidcSecretName is blank, default=" + DefaultPlatformOidcSecretName)
		platformOidcSecretName = DefaultPlatformOidcSecretName
	}

	cfcRouterURL := "https://icp-management-ingress." + ingressNamespace + ":443"
	commonClusterEnvVars := BuildCommonClusterEnvVars(instanceNamespace, iamNamespace)

	uiClusterEnvVars := []corev1.EnvVar{
		{
			Name:  "COMMON_HEADER_NAMESPACE",
			Value: headerNamespace,
		},
		{
			Name:  "cfcRouterUrl",
			Value: cfcRouterURL,
		},
		{
			Name: "ICP_API_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: apiKeySecretName,
					},
					Key:      "ICP_API_KEY",
					Optional: &TrueVar,
				},
			},
		},
		{
			Name: "WLP_CLIENT_ID",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: platformOidcSecretName,
					},
					Key:      "WLP_CLIENT_ID",
					Optional: &TrueVar,
				},
			},
		},
		{
			Name: "WLP_CLIENT_SECRET",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: platformOidcSecretName,
					},
					Key:      "WLP_CLIENT_SECRET",
					Optional: &TrueVar,
				},
			},
		},
	}

	clusterEnvVars := append(commonClusterEnvVars, uiClusterEnvVars...)

	var providerURL string
	if isMcmUI {
		providerURL = cfcRouterURL + "/idprovider"
	} else {
		providerURL = "https://platform-identity-provider." + iamNamespace + ":4300"

		var clusterName string
		if instanceClusterName != "" {
			clusterName = instanceClusterName
		} else {
			clusterName = DefaultClusterName
		}
		nameEnvVar := corev1.EnvVar{
			Name:  "CLUSTER_NAME",
			Value: clusterName,
		}
		clusterEnvVars = append(clusterEnvVars, nameEnvVar)
	}
	providerEnvVar := corev1.EnvVar{
		Name:  "PLATFORM_IDENTITY_PROVIDER_URL",
		Value: providerURL,
	}
	clusterEnvVars = append(clusterEnvVars, providerEnvVar)
	return clusterEnvVars
}

func BuildSenderClusterEnvVars(instanceNamespace, instanceClusterNamespace,
	instanceClusterName, hubKubeConfigSecret string) []corev1.EnvVar {

	reqLogger := log.WithValues("func", "BuildSenderClusterEnvVars")

	var clusterName string
	if instanceClusterName != "" {
		clusterName = instanceClusterName
	} else {
		clusterName = DefaultClusterName
	}

	var clusterNamespace string
	if instanceClusterNamespace != "" {
		reqLogger.Info("clusterNamespace=" + instanceClusterNamespace)
		clusterNamespace = instanceClusterNamespace
	} else {
		reqLogger.Info("clusterNamespace is blank, use instance=" + instanceNamespace)
		clusterNamespace = instanceNamespace
	}

	clusterEnvVars := []corev1.EnvVar{
		{
			Name:  "HC_CLUSTER_NAME",
			Value: clusterName,
		},
		{
			Name:  "HC_CLUSTER_NAMESPACE",
			Value: clusterNamespace,
		},
		{
			Name:  "HC_HUB_CONFIG",
			Value: hubKubeConfigSecret,
		},
	}

	return clusterEnvVars
}

func BuildReceiverEnvVars(multiCloudReceiverEnabled bool) []corev1.EnvVar {
	receiverEnvVars := []corev1.EnvVar{
		{
			Name:  "HC_DM_MCM_RECEIVER_ENABLED",
			Value: strconv.FormatBool(multiCloudReceiverEnabled),
		},
	}
	if multiCloudReceiverEnabled {
		receiverEnvVars = append(receiverEnvVars, ReceiverSslEnvVars...)
	}
	return receiverEnvVars
}

// set loglevelType to "log4js" when building volumes for metering-mcmui.
// set loglevelType to "loglevel" when building volumes for any other component.
func BuildCommonVolumes(mongoDB operatorv1alpha1.MeteringSpecMongoDB, loglevelPrefix, loglevelType string) []corev1.Volume {
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
					SecretName:  mongoDB.ClusterCertsSecret,
					DefaultMode: &DefaultMode,
					Optional:    &TrueVar,
				},
			},
		},
		{
			Name: "mongodb-client-cert",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  mongoDB.ClientCertsSecret,
					DefaultMode: &DefaultMode,
					Optional:    &TrueVar,
				},
			},
		},
		{
			Name: "muser-icp-mongodb-admin",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  mongoDB.UsernameSecret,
					DefaultMode: &DefaultMode,
					Optional:    &TrueVar,
				},
			},
		},
		{
			Name: "mpass-icp-mongodb-admin",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  mongoDB.PasswordSecret,
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

func BuildUISecretVolumes(apiKeySecretName, platformOidcSecretName string) []corev1.Volume {
	uiVolumes := []corev1.Volume{
		{
			Name: apiKeySecretName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  apiKeySecretName,
					DefaultMode: &DefaultMode,
					Optional:    &TrueVar,
				},
			},
		},
		{
			Name: platformOidcSecretName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  platformOidcSecretName,
					DefaultMode: &DefaultMode,
					Optional:    &TrueVar,
				},
			},
		},
	}
	return uiVolumes
}

func BuildUISecretVolumeMounts(apiKeySecretName, platformOidcSecretName string) []corev1.VolumeMount {
	uiVolumeMounts := []corev1.VolumeMount{
		{
			Name:      apiKeySecretName,
			MountPath: "/sec/" + apiKeySecretName,
		},
		{
			Name:      platformOidcSecretName,
			MountPath: "/sec/" + platformOidcSecretName,
		},
	}
	return uiVolumeMounts
}

// checkerCommand is the command to be executed by the secret-check container.
// mongoDB contains the password names from the CR.
// additionalInfo contains info about additional secrets to check.
func BuildSecretCheckContainer(deploymentName, imageName, checkerCommand string,
	mongoDB operatorv1alpha1.MeteringSpecMongoDB, additionalInfo *SecretCheckData) corev1.Container {

	containerName := deploymentName + "-secret-check"
	nameList := mongoDB.UsernameSecret + " " + mongoDB.PasswordSecret + " " +
		mongoDB.ClusterCertsSecret + " " + mongoDB.ClientCertsSecret
	usernameSecretDir := "muser-" + mongoDB.UsernameSecret
	passwordSecretDir := "mpass-" + mongoDB.PasswordSecret
	dirList := usernameSecretDir + " " + passwordSecretDir + " " +
		mongoDB.ClusterCertsSecret + " " + mongoDB.ClientCertsSecret
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "mongodb-ca-cert",
			MountPath: "/sec/" + mongoDB.ClusterCertsSecret,
		},
		{
			Name:      "mongodb-client-cert",
			MountPath: "/sec/" + mongoDB.ClientCertsSecret,
		},
		{
			Name:      usernameSecretDir,
			MountPath: "/sec/" + usernameSecretDir,
		},
		{
			Name:      passwordSecretDir,
			MountPath: "/sec/" + passwordSecretDir,
		},
	}
	if additionalInfo != nil {
		nameList += " "
		nameList += additionalInfo.Names
		dirList += " "
		dirList += additionalInfo.Dirs
		volumeMounts = append(volumeMounts, additionalInfo.VolumeMounts...)
	}

	var secretCheckContainer = corev1.Container{
		Image:           imageName,
		Name:            containerName,
		ImagePullPolicy: corev1.PullAlways,
		Command: []string{
			"sh",
			"-c",
			checkerCommand,
		},
		Env: []corev1.EnvVar{
			{
				Name:  "SECRET_LIST",
				Value: nameList,
			},
			{
				Name:  "SECRET_DIR_LIST",
				Value: dirList,
			},
		},
		VolumeMounts:    volumeMounts,
		Resources:       commonInitResources,
		SecurityContext: &commonSecurityContext,
	}
	return secretCheckContainer
}

func BuildInitContainer(deploymentName, imageName string, envVars []corev1.EnvVar) corev1.Container {
	containerName := deploymentName + "-init"
	var initContainer = corev1.Container{
		Image:           imageName,
		Name:            containerName,
		ImagePullPolicy: corev1.PullAlways,
		Command: []string{
			"node",
			"/datamanager/lib/metering_init.js",
			"verifyOnlyMongo",
		},
		// CommonEnvVars and mongoDBEnvVars will be added by the controller
		Env:             envVars,
		VolumeMounts:    commonInitVolumeMounts,
		Resources:       commonInitResources,
		SecurityContext: &commonSecurityContext,
	}
	return initContainer
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

// returns the annotations associated with the pod being created
func AnnotationsForPod() map[string]string {
	return map[string]string{"productName": MeteringProductName,
		"productID": MeteringProductID, "productVersion": MeteringProductVersion}
}

// GetPodNames returns the pod names of the array of pods passed in
func GetPodNames(pods []corev1.Pod) []string {
	reqLogger := log.WithValues("func", "GetPodNames")
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
		reqLogger.Info("pod name=" + pod.Name)
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
