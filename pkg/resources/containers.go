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

// Linter doesn't like "Secret" in string var names that are assigned a value,
// so use concatenation to create the value.
// Example:  const MySecretName = "metering-secret" + ""

package resources

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// SecretCheckData contains info about additional secrets for the secret-check container.
// Names will be added to the SECRET_LIST env var.
// Dirs will be added to the SECRET_DIR_LIST env var.
// VolumeMounts contains the volume mounts associated with the secrets.
type SecretCheckData struct {
	Names        string
	Dirs         string
	VolumeMounts []corev1.VolumeMount
}

const DefaultImageRegistry = "quay.io/opencloudio"
const DefaultDmImageName = "metering-data-manager"
const DefaultReportImageName = "metering-report"
const DefaultUIImageName = "metering-ui"
const DefaultMcmUIImageName = "metering-mcmui"
const DefaultClusterIssuer = "cs-ca-clusterissuer"
const DefaultAPIServiceName = "v1.metering.ibm.com"
const DefaultWatchNamespace = "ibm-common-services"
const WatchNamespaceV350 = "ibm-common-services"

// starting with Common Services 3.4, images can be pulled by SHA or tag.
// run scripts/get-image-sha.sh to update operator.yaml with the SHA values.
// a SHA value looks like this: "sha256:nnnnnnnn"
// a tag value looks like this: "3.5.0".
const DefaultDmImageTag = "3.6.1"
const DefaultReportImageTag = "3.6.1"
const DefaultUIImageTag = "3.6.1"
const DefaultMcmUIImageTag = "3.6.1"

// define the env vars that contain either the SHA or the tag
const VarImageSHAforDM = "IMAGE_SHA_OR_TAG_DM"
const VarImageSHAforUI = "IMAGE_SHA_OR_TAG_UI"
const VarImageSHAforMCMUI = "IMAGE_SHA_OR_TAG_MCMUI"
const VarImageSHAforReport = "IMAGE_SHA_OR_TAG_REPORT"

// use concatenation so linter won't complain about "Secret" vars
const DefaultAPIKeySecretName = "icp-serviceid-apikey-secret" + ""
const DefaultPlatformOidcSecretName = "platform-oidc-credentials" + ""

var TrueVar = true
var FalseVar = false
var Replica1 int32 = 1
var Seconds60 int64 = 60

var cpu100 = resource.NewMilliQuantity(100, resource.DecimalSI)          // 100m
var cpu500 = resource.NewMilliQuantity(500, resource.DecimalSI)          // 500m
var cpu1000 = resource.NewMilliQuantity(1000, resource.DecimalSI)        // 1000m
var memory100 = resource.NewQuantity(100*1024*1024, resource.BinarySI)   // 100Mi
var memory128 = resource.NewQuantity(128*1024*1024, resource.BinarySI)   // 128Mi
var memory256 = resource.NewQuantity(256*1024*1024, resource.BinarySI)   // 256Mi
var memory512 = resource.NewQuantity(512*1024*1024, resource.BinarySI)   // 512Mi
var memory2560 = resource.NewQuantity(2560*1024*1024, resource.BinarySI) // 2560Mi

const DefaultClusterName = "mycluster"

var ArchitectureList = []string{
	"amd64",
	"ppc64le",
	"s390x",
}

var CommonEnvVars = []corev1.EnvVar{
	{
		Name:  "NODE_TLS_REJECT_UNAUTHORIZED",
		Value: "0",
	},
}

var IAMEnvVars = []corev1.EnvVar{
	{
		Name:  "DEFAULT_IAM_TOKEN_SERVICE_PORT",
		Value: "10443",
	},
	{
		Name:  "DEFAULT_IAM_PAP_SERVICE_PORT",
		Value: "39001",
	},
}

var SecretCheckCmd = `set -- $SECRET_LIST; ` +
	`for secretDirName in $SECRET_DIR_LIST; do` +
	`  while true; do` +
	`    echo ` + "`date`" + `: Checking for secret $1;` +
	`    ls /sec/$secretDirName/* && break;` +
	`    echo ` + "`date`" + `: Required secret $1 not found ... try again in 30s;` +
	`    sleep 30;` +
	`  done;` +
	`  echo ` + "`date`" + `: Secret $1 found;` +
	`  shift; ` +
	`done; ` +
	`echo ` + "`date`" + `: All required secrets exist`

var SenderSecretCheckCmd = SecretCheckCmd + ";" +
	`echo ` + "`date`" + `: Further, checking for kubeConfig secret...;` +
	`node /datamanager/lib/metering_init.js kubeconfig_secretcheck `

// API certificate definition
const APICertName = "icp-metering-api-ca-cert"
const APICertCommonName = "metering-server"
const APICertDirName = "metering-api"

// use concatenation so linter won't complain about "Secret" vars
const APICertSecretName = "icp-metering-api-secret" + ""
const APICertVolumeName = "icp-metering-api-certs"

var APICertVolumeMount = corev1.VolumeMount{
	Name:      APICertVolumeName,
	MountPath: "/sec/" + APICertDirName,
}

var APICertVolume = corev1.Volume{
	Name: APICertVolumeName,
	VolumeSource: corev1.VolumeSource{
		Secret: &corev1.SecretVolumeSource{
			SecretName:  APICertSecretName,
			DefaultMode: &DefaultMode,
			Optional:    &TrueVar,
		},
	},
}

// UI certificate definition
const UICertName = "metering-ui-ca-cert"
const UICertCommonName = "metering-ui"
const UICertDirName = "metering-ui"

// use concatenation so linter won't complain about "Secret" vars
const UICertSecretName = "metering-ui-cert" + ""
const UICertVolumeName = "metering-ui-certs"

var UICertVolumeMountForSecretCheck = corev1.VolumeMount{
	Name:      UICertVolumeName,
	MountPath: "/sec/" + UICertDirName,
}
var UICertVolumeMountForMain = corev1.VolumeMount{
	Name:      UICertVolumeName,
	MountPath: "/certs/" + UICertDirName,
}

var UICertVolume = corev1.Volume{
	Name: UICertVolumeName,
	VolumeSource: corev1.VolumeSource{
		Secret: &corev1.SecretVolumeSource{
			SecretName:  UICertSecretName,
			DefaultMode: &DefaultMode,
			Optional:    &TrueVar,
		},
	},
}

// MCMUI certificate definition
const McmUICertName = "metering-mcmui-ca-cert"
const McmUICertCommonName = "metering-mcmui"
const McmUICertDirName = "metering-mcmui"

// use concatenation so linter won't complain about "Secret" vars
const McmUICertSecretName = "metering-mcmui-cert" + ""
const McmUICertVolumeName = "metering-mcmui-certs"

var McmUICertVolumeMountForSecretCheck = corev1.VolumeMount{
	Name:      McmUICertVolumeName,
	MountPath: "/sec/" + McmUICertDirName,
}
var McmUICertVolumeMountForMain = corev1.VolumeMount{
	Name:      McmUICertVolumeName,
	MountPath: "/certs/" + McmUICertDirName,
}

var McmUICertVolume = corev1.Volume{
	Name: McmUICertVolumeName,
	VolumeSource: corev1.VolumeSource{
		Secret: &corev1.SecretVolumeSource{
			SecretName:  McmUICertSecretName,
			DefaultMode: &DefaultMode,
			Optional:    &TrueVar,
		},
	},
}

// Receiver certificate definition
const ReceiverCertName = "icp-metering-receiver-ca-cert"
const ReceiverCertCommonName = "metering-receiver"
const ReceiverCertDirName = "metering-receiver"

// use concatenation so linter won't complain about "Secret" vars
const ReceiverCertSecretName = "icp-metering-receiver-secret" + ""
const ReceiverCertVolumeName = "icp-metering-receiver-certs"

var ReceiverCertVolumeMountForSecretCheck = corev1.VolumeMount{
	Name:      ReceiverCertVolumeName,
	MountPath: "/sec/" + ReceiverCertDirName,
}
var ReceiverCertVolumeMountForMain = corev1.VolumeMount{
	Name:      ReceiverCertVolumeName,
	MountPath: "/certs/" + ReceiverCertDirName,
}
var ReceiverCertVolume = corev1.Volume{
	Name: ReceiverCertVolumeName,
	VolumeSource: corev1.VolumeSource{
		Secret: &corev1.SecretVolumeSource{
			SecretName:  ReceiverCertSecretName,
			DefaultMode: &DefaultMode,
			Optional:    &TrueVar,
		},
	},
}

//***************************************************************
// these vars are used when multiCloudReceiverEnabled = true
//***************************************************************

var ReceiverSslEnvVars = []corev1.EnvVar{
	{
		Name:  "HC_RECEIVER_SSL_CA",
		Value: "/certs/" + ReceiverCertDirName + "/ca.crt",
	},
	{
		Name:  "HC_RECEIVER_SSL_CERT",
		Value: "/certs/" + ReceiverCertDirName + "/tls.crt",
	},
	{
		Name:  "HC_RECEIVER_SSL_KEY",
		Value: "/certs/" + ReceiverCertDirName + "/tls.key",
	},
}

//***************************************************************
// Common definitions
//***************************************************************
var commonInitVolumeMounts = []corev1.VolumeMount{
	{
		Name:      "mongodb-ca-cert",
		MountPath: "/certs/mongodb-ca",
	},
	{
		Name:      "mongodb-client-cert",
		MountPath: "/certs/mongodb-client",
	},
}

var commonInitResources = corev1.ResourceRequirements{
	Limits: map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    *cpu100,
		corev1.ResourceMemory: *memory100},
	Requests: map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    *cpu100,
		corev1.ResourceMemory: *memory100},
}

var commonSecurityContext = corev1.SecurityContext{
	AllowPrivilegeEscalation: &FalseVar,
	Privileged:               &FalseVar,
	ReadOnlyRootFilesystem:   &TrueVar,
	RunAsNonRoot:             &TrueVar,
	Capabilities: &corev1.Capabilities{
		Drop: []corev1.Capability{
			"ALL",
		},
	},
}

var CommonMainVolumeMounts = []corev1.VolumeMount{
	{
		Name:      "mongodb-ca-cert",
		MountPath: "/certs/mongodb-ca",
	},
	{
		Name:      "mongodb-client-cert",
		MountPath: "/certs/mongodb-client",
	},
}

var LoglevelVolumeMount = corev1.VolumeMount{
	Name:      "loglevel",
	MountPath: "/etc/config",
}

var Log4jsVolumeMount = corev1.VolumeMount{
	Name:      "log4js",
	MountPath: "/etc/config",
}

var TempDirVolume = corev1.Volume{
	Name: "tmp-dir",
	VolumeSource: corev1.VolumeSource{
		EmptyDir: &corev1.EmptyDirVolumeSource{},
	},
}

//***************************************************************
// DataManager container definition
//***************************************************************

var DmResourceRequirements = corev1.ResourceRequirements{
	Limits: map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    *cpu1000,
		corev1.ResourceMemory: *memory2560},
	Requests: map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    *cpu100,
		corev1.ResourceMemory: *memory256},
}

var DmMainContainer = corev1.Container{
	Image:           "metering-data-manager",
	Name:            "metering-dm",
	ImagePullPolicy: corev1.PullAlways,
	// CommonMainVolumeMounts will be added by the controller
	VolumeMounts: []corev1.VolumeMount{
		LoglevelVolumeMount,
	},
	// CommonEnvVars, IAMEnvVars and mongoDBEnvVars will be added by the controller.
	// HC_DM_MCM_RECEIVER_ENABLED will be set by BuildReceiverEnvVars().
	// Removed ICP_API_KEY.
	Env: []corev1.EnvVar{
		{
			Name:  "METERING_API_ENABLED",
			Value: "true",
		},
		{
			Name:  "HC_DM_USE_HTTPS",
			Value: "false",
		},
		{
			Name:  "HC_DM_MCM_SENDER_ENABLED",
			Value: "false",
		},
		{
			Name:  "HC_DM_STORAGEREADER_ENABLED",
			Value: "true",
		},
		{
			Name:  "HC_DM_REPORTER2_ENABLED",
			Value: "true",
		},
		{
			Name:  "HC_DM_PURGER2_ENABLED",
			Value: "true",
		},
		{
			Name:  "HC_DM_PREAGGREGATOR_ENABLED",
			Value: "true",
		},
		{
			Name:  "HC_DM_METRICS_ENABLED",
			Value: "true",
		},
		{
			Name:  "HC_DM_SELFMETER_PURGER_ENABLED",
			Value: "true",
		},
	},
	Ports: []corev1.ContainerPort{
		{ContainerPort: 3000},
		{ContainerPort: 5000},
	},
	LivenessProbe: &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/livenessProbe",
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 3000,
				},
				Scheme: corev1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: 305,
		TimeoutSeconds:      5,
		PeriodSeconds:       300,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	},
	ReadinessProbe: &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/readinessProbe",
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 3000,
				},
				Scheme: corev1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: 15,
		TimeoutSeconds:      15,
		PeriodSeconds:       30,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	},
	Resources:       corev1.ResourceRequirements{},
	SecurityContext: &commonSecurityContext,
}

//***************************************************************
// Report Server container definition
//***************************************************************

var ReportResourceRequirements = corev1.ResourceRequirements{
	Limits: map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    *cpu500,
		corev1.ResourceMemory: *memory512},
	Requests: map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    *cpu100,
		corev1.ResourceMemory: *memory128},
}

var ReportContainer = corev1.Container{
	Image:           "metering-report",
	Name:            "metering-report",
	ImagePullPolicy: corev1.PullAlways,
	Command: []string{
		"./apiserver",
	},
	Args: []string{
		"--cert-dir=/tmp",
		"--secure-port=7443",
		"--disable-admission-plugins=NamespaceLifecycle,MutatingAdmissionWebhook,ValidatingAdmissionWebhook",
	},
	VolumeMounts: []corev1.VolumeMount{
		{
			Name:      "tmp-dir",
			MountPath: "/tmp",
		},
		{
			Name:      APICertVolumeName,
			MountPath: "/certs/" + APICertDirName,
		},
	},
	SecurityContext: &commonSecurityContext,
	LivenessProbe: &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/livenessProbe",
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 3000,
				},
				Scheme: corev1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: 305,
		TimeoutSeconds:      5,
		PeriodSeconds:       300,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	},
	ReadinessProbe: &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/readinessProbe",
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 3000,
				},
				Scheme: corev1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: 15,
		TimeoutSeconds:      15,
		PeriodSeconds:       30,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	},
	Resources: corev1.ResourceRequirements{},
}

//***************************************************************
// Reader container definition
//***************************************************************

var RdrResourceRequirements = corev1.ResourceRequirements{
	Limits: map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    *cpu500,
		corev1.ResourceMemory: *memory512},
	Requests: map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    *cpu100,
		corev1.ResourceMemory: *memory128},
}

var RdrMainContainer = corev1.Container{
	Image:           "metering-data-manager",
	Name:            "metering-reader",
	ImagePullPolicy: corev1.PullAlways,
	// CommonMainVolumeMounts will be added by the controller
	VolumeMounts: []corev1.VolumeMount{
		{
			Name:      APICertVolumeName,
			MountPath: "/certs/" + APICertDirName,
		},
		LoglevelVolumeMount,
	},
	// CommonEnvVars, IAMEnvVars and mongoDBEnvVars will be added by the controller.
	// Removed ICP_API_KEY.
	Env: []corev1.EnvVar{
		{
			Name:  "METERING_API_ENABLED",
			Value: "true",
		},
		{
			Name:  "METERING_INTERNALAPI_ENABLED",
			Value: "true",
		},
		{
			Name:  "HC_DM_USE_HTTPS",
			Value: "true",
		},
		{
			Name:  "HC_DM_MCM_RECEIVER_ENABLED",
			Value: "false",
		},
		{
			Name:  "HC_DM_MCM_SENDER_ENABLED",
			Value: "false",
		},
		{
			Name:  "HC_DM_MCMREADER_ENABLED",
			Value: "false",
		},
		{
			Name:  "HC_DM_READER_ENABLED",
			Value: "true",
		},
		{
			Name:  "HC_DM_STORAGEREADER_ENABLED",
			Value: "false",
		},
		{
			Name:  "HC_DM_READER_APIENABLED",
			Value: "true",
		},
		{
			Name:  "HC_DM_REPORTER2_ENABLED",
			Value: "false",
		},
		{
			Name:  "HC_DM_PURGER2_ENABLED",
			Value: "false",
		},
		{
			Name:  "HC_DM_PREAGGREGATOR_ENABLED",
			Value: "false",
		},
		{
			Name:  "HC_DM_SELFMETER_PURGER_ENABLED",
			Value: "false",
		},
		{
			Name:  "HC_DM_API_PORT",
			Value: "4000",
		},
		{
			Name:  "HC_DM_INTERNAL_API_PORT",
			Value: "4002",
		},
		{
			Name:  "HC_API_SSL_CA",
			Value: "/certs/" + APICertDirName + "/ca.crt",
		},
		{
			Name:  "HC_API_SSL_CERT",
			Value: "/certs/" + APICertDirName + "/tls.crt",
		},
		{
			Name:  "HC_API_SSL_KEY",
			Value: "/certs/" + APICertDirName + "/tls.key",
		},
	},
	Ports: []corev1.ContainerPort{
		{ContainerPort: 3000},
		{ContainerPort: 4000},
		{ContainerPort: 4002},
	},
	LivenessProbe: &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/livenessProbe",
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 3000,
				},
				Scheme: corev1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: 305,
		TimeoutSeconds:      5,
		PeriodSeconds:       300,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	},
	ReadinessProbe: &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/readinessProbe",
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 3000,
				},
				Scheme: corev1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: 15,
		TimeoutSeconds:      15,
		PeriodSeconds:       30,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	},
	Resources:       corev1.ResourceRequirements{},
	SecurityContext: &commonSecurityContext,
}

//***************************************************************
// Sender container definition
//***************************************************************

var SenderResourceRequirements = corev1.ResourceRequirements{
	Limits: map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    *cpu500,
		corev1.ResourceMemory: *memory512},
	Requests: map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    *cpu100,
		corev1.ResourceMemory: *memory128},
}

var SenderMainContainer = corev1.Container{
	Image:           "metering-data-manager",
	Name:            "metering-sender",
	ImagePullPolicy: corev1.PullAlways,
	// CommonMainVolumeMounts will be added by the controller
	VolumeMounts: []corev1.VolumeMount{
		LoglevelVolumeMount,
	},
	// CommonEnvVars, IAMEnvVars and mongoDBEnvVars will be added by the controller
	Env: []corev1.EnvVar{
		{
			Name:  "METERING_API_ENABLED",
			Value: "false",
		},
		{
			Name:  "HC_DM_SELFMETER_PURGER_ENABLED",
			Value: "false",
		},
		{
			Name:  "HC_DM_REPORTER2_ENABLED",
			Value: "false",
		},
		{
			Name:  "HC_DM_PURGER2_ENABLED",
			Value: "false",
		},
		{
			Name:  "HC_DM_PREAGGREGATOR_ENABLED",
			Value: "false",
		},
		{
			Name:  "HC_DM_METRICS_ENABLED",
			Value: "false",
		},
		{
			Name:  "HC_DM_READER_APIENABLED",
			Value: "false",
		},
		{
			Name:  "HC_DM_MCM_RECEIVER_ENABLED",
			Value: "false",
		},
		{
			Name:  "HC_DM_MCMREADER_ENABLED",
			Value: "false",
		},
		{
			Name:  "HC_DM_MCM_SENDER_ENABLED",
			Value: "true",
		},
	},
	LivenessProbe: &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/livenessProbe",
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 3000,
				},
				Scheme: corev1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: 305,
		TimeoutSeconds:      5,
		PeriodSeconds:       300,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	},
	ReadinessProbe: &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/readinessProbe",
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 3000,
				},
				Scheme: corev1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: 15,
		TimeoutSeconds:      15,
		PeriodSeconds:       30,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	},
	Resources:       corev1.ResourceRequirements{},
	SecurityContext: &commonSecurityContext,
}

var UIEnvVars = []corev1.EnvVar{
	{
		Name:  "IS_PRIVATECLOUD",
		Value: "true",
	},
	{
		Name:  "USE_PRIVATECLOUD_SECURITY",
		Value: "true",
	},
	{
		Name:  "DEFAULT_PLATFORM_IDENTITY_MANAGEMENT_SERVICE_PORT",
		Value: "4500",
	},
	{
		Name:  "DEFAULT_PLATFORM_HEADER_SERVICE_PORT",
		Value: "3000",
	},
}

//***************************************************************
// UI container definition
//***************************************************************

var UIResourceRequirements = corev1.ResourceRequirements{
	Limits: map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    *cpu500,
		corev1.ResourceMemory: *memory512},
	Requests: map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    *cpu100,
		corev1.ResourceMemory: *memory128},
}

var UIMainContainer = corev1.Container{
	Image:           "metering-ui",
	Name:            "metering-ui",
	ImagePullPolicy: corev1.PullAlways,
	// CommonMainVolumeMounts will be added by the controller
	VolumeMounts: []corev1.VolumeMount{
		LoglevelVolumeMount,
	},
	// CommonEnvVars, IAMEnvVars, UIEnvVars and mongoDBEnvVars will be added by the controller. removed HC_HCAPI_URI
	Env: []corev1.EnvVar{
		{
			Name:  "ICP_DEFAULT_DASHBOARD",
			Value: "cpi.icp.main",
		},
		{
			Name:  "PORT",
			Value: "3130",
		},
		{
			Name:  "PROXY_URI",
			Value: "metering",
		},
		{
			Name:  "HC_UI_ISSSL",
			Value: "true",
		},
		{
			Name:  "HC_UI_SSL_CA",
			Value: "/certs/" + UICertDirName + "/ca.crt",
		},
		{
			Name:  "HC_UI_SSL_CERT",
			Value: "/certs/" + UICertDirName + "/tls.crt",
		},
		{
			Name:  "HC_UI_SSL_KEY",
			Value: "/certs/" + UICertDirName + "/tls.key",
		},
	},
	Ports: []corev1.ContainerPort{
		{ContainerPort: 3130},
	},
	LivenessProbe: &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/unsecure/c2c/status/livenessProbe",
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 3130,
				},
				Scheme: corev1.URISchemeHTTPS,
			},
		},
		InitialDelaySeconds: 305,
		TimeoutSeconds:      5,
		PeriodSeconds:       300,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	},
	ReadinessProbe: &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/unsecure/c2c/status/readinessProbe",
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 3130,
				},
				Scheme: corev1.URISchemeHTTPS,
			},
		},
		InitialDelaySeconds: 15,
		TimeoutSeconds:      15,
		PeriodSeconds:       30,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	},
	Resources:       corev1.ResourceRequirements{},
	SecurityContext: &commonSecurityContext,
}

//***************************************************************
// MCM UI container definition
//***************************************************************

var McmUIResourceRequirements = corev1.ResourceRequirements{
	Limits: map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    *cpu500,
		corev1.ResourceMemory: *memory256},
	Requests: map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    *cpu100,
		corev1.ResourceMemory: *memory128},
}

var McmUIMainContainer = corev1.Container{
	Image:           "metering-mcmui",
	Name:            "metering-mcmui",
	ImagePullPolicy: corev1.PullAlways,
	// CommonMainVolumeMounts will be added by the controller
	VolumeMounts: []corev1.VolumeMount{
		Log4jsVolumeMount,
	},
	// CommonEnvVars, IAMEnvVars, UIEnvVars and mongoDBEnvVars will be added by the controller
	Env: []corev1.EnvVar{
		{
			Name:  "PORT",
			Value: "3001",
		},
		{
			Name:  "PROXY_URI",
			Value: "metering-mcm",
		},
		{
			Name:  "MCM_UI_ISSSL",
			Value: "true",
		},
		{
			Name:  "MCM_UI_SSL_CA",
			Value: "/certs/" + McmUICertDirName + "/ca.crt",
		},
		{
			Name:  "MCM_UI_SSL_CERT",
			Value: "/certs/" + McmUICertDirName + "/tls.crt",
		},
		{
			Name:  "MCM_UI_SSL_KEY",
			Value: "/certs/" + McmUICertDirName + "/tls.key",
		},
	},
	Ports: []corev1.ContainerPort{
		{ContainerPort: 3001},
	},
	LivenessProbe: &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/unsecure/livenessProbe",
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 3001,
				},
				Scheme: corev1.URISchemeHTTPS,
			},
		},
		InitialDelaySeconds: 305,
		TimeoutSeconds:      5,
		PeriodSeconds:       300,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	},
	ReadinessProbe: &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/unsecure/readinessProbe",
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 3001,
				},
				Scheme: corev1.URISchemeHTTPS,
			},
		},
		InitialDelaySeconds: 15,
		TimeoutSeconds:      5,
		PeriodSeconds:       15,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	},
	Resources:       corev1.ResourceRequirements{},
	SecurityContext: &commonSecurityContext,
}
