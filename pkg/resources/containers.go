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
const DefaultDmImageTag = "3.4.0"
const DefaultReportImageName = "metering-report"
const DefaultReportImageTag = "3.4.0"
const DefaultUIImageName = "metering-ui"
const DefaultUIImageTag = "3.4.0"
const DefaultMcmUIImageName = "metering-mcmui"
const DefaultMcmUIImageTag = "3.4.0"
const DefaultSenderImageName = "metering-data-manager"
const DefaultSenderImageTag = "3.4.0"
const DefaultClusterIssuer = "cs-ca-clusterissuer"
const DefaultAPIServiceName = "v1.metering.ibm.com"

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

const APICertName = "icp-metering-api-ca-cert"
const APICertCommonName = "metering-server"

// use concatenation so linter won't complain about "Secret" vars
const APICertSecretName = "icp-metering-api-secret" + ""
const APICertVolumeName = "icp-metering-api-certs"

var APICertVolumeMount = corev1.VolumeMount{
	Name:      APICertVolumeName,
	MountPath: "/sec/" + APICertSecretName,
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

var TempDirVolume = corev1.Volume{
	Name: "tmp-dir",
	VolumeSource: corev1.VolumeSource{
		EmptyDir: &corev1.EmptyDirVolumeSource{},
	},
}

const ReceiverCertName = "icp-metering-receiver-ca-cert"
const ReceiverCertCommonName = "metering-receiver"

// use concatenation so linter won't complain about "Secret" vars
const ReceiverCertSecretName = "icp-metering-receiver-secret" + ""
const ReceiverCertVolumeName = "icp-metering-receiver-certs"

var ReceiverCertVolumeMountForSecretCheck = corev1.VolumeMount{
	Name:      ReceiverCertVolumeName,
	MountPath: "/sec/" + ReceiverCertSecretName,
}
var ReceiverCertVolumeMountForMain = corev1.VolumeMount{
	Name:      ReceiverCertVolumeName,
	MountPath: "/certs/" + ReceiverCertCommonName,
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

var ReceiverSslEnvVars = []corev1.EnvVar{
	{
		Name:  "HC_RECEIVER_SSL_CA",
		Value: "/certs/" + ReceiverCertCommonName + "/ca.crt",
	},
	{
		Name:  "HC_RECEIVER_SSL_CERT",
		Value: "/certs/" + ReceiverCertCommonName + "/tls.crt",
	},
	{
		Name:  "HC_RECEIVER_SSL_KEY",
		Value: "/certs/" + ReceiverCertCommonName + "/tls.key",
	},
}

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
		{
			Name:  "HC_DM_IS_ICP",
			Value: "true",
		},
		{
			Name:  "HC_DM_ALLOW_TEST",
			Value: "false",
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
	Resources: corev1.ResourceRequirements{
		Limits: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    *cpu1000,
			corev1.ResourceMemory: *memory2560},
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    *cpu100,
			corev1.ResourceMemory: *memory256},
	},
	SecurityContext: &commonSecurityContext,
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
	},
	// CommonMainVolumeMounts will be added by the controller
	VolumeMounts: []corev1.VolumeMount{
		{
			Name:      "tmp-dir",
			MountPath: "/tmp",
		},
		{
			Name:      APICertVolumeName,
			MountPath: "/certs/metering-api",
		},
	},
	SecurityContext: &commonSecurityContext,
	Env: []corev1.EnvVar{
		{
			Name:  "SA_NAME",
			Value: "ibm-metering-operator",
		},
	},
}

var RdrMainContainer = corev1.Container{
	Image:           "metering-data-manager",
	Name:            "metering-reader",
	ImagePullPolicy: corev1.PullAlways,
	// CommonMainVolumeMounts will be added by the controller
	VolumeMounts: []corev1.VolumeMount{
		{
			Name:      APICertVolumeName,
			MountPath: "/certs/metering-api",
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
			Value: "false",
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
			Name:  "HC_DM_IS_ICP",
			Value: "true",
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
			Value: "/certs/metering-api/ca.crt",
		},
		{
			Name:  "HC_API_SSL_CERT",
			Value: "/certs/metering-api/tls.crt",
		},
		{
			Name:  "HC_API_SSL_KEY",
			Value: "/certs/metering-api/tls.key",
		},
		{
			Name: "MY_NODE_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath:  "spec.nodeName",
					APIVersion: "v1",
				},
			},
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
	Resources: corev1.ResourceRequirements{
		Limits: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    *cpu500,
			corev1.ResourceMemory: *memory512},
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    *cpu100,
			corev1.ResourceMemory: *memory128},
	},
	SecurityContext: &commonSecurityContext,
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
		{
			Name:  "HC_DM_IS_ICP",
			Value: "true",
		},
		{
			Name:  "HC_DM_ALLOW_TEST",
			Value: "false",
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
	Resources: corev1.ResourceRequirements{
		Limits: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    *cpu500,
			corev1.ResourceMemory: *memory512},
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    *cpu100,
			corev1.ResourceMemory: *memory128},
	},
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
	{
		Name:  "HC_DM_ALLOW_TEST",
		Value: "false",
	},
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
				Path: "/unsecure/c2c/status/readinessProbe",
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 3130,
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
	Resources: corev1.ResourceRequirements{
		Limits: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    *cpu500,
			corev1.ResourceMemory: *memory512},
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    *cpu100,
			corev1.ResourceMemory: *memory128},
	},
	SecurityContext: &commonSecurityContext,
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
				Path: "/unsecure/readinessProbe",
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 3001,
				},
				Scheme: corev1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: 15,
		TimeoutSeconds:      5,
		PeriodSeconds:       15,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	},
	Resources: corev1.ResourceRequirements{
		Limits: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    *cpu500,
			corev1.ResourceMemory: *memory256},
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    *cpu100,
			corev1.ResourceMemory: *memory128},
	},
	SecurityContext: &commonSecurityContext,
}
