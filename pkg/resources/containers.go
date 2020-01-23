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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var trueVar = true
var falseVar = false

//CS??? var user99 int64 = 99
var cpu100 = resource.NewMilliQuantity(100, resource.DecimalSI)          // 100m
var cpu500 = resource.NewMilliQuantity(500, resource.DecimalSI)          // 500m
var cpu1000 = resource.NewMilliQuantity(1000, resource.DecimalSI)        // 1000m
var memory100 = resource.NewQuantity(100*1024*1024, resource.BinarySI)   // 100Mi
var memory128 = resource.NewQuantity(128*1024*1024, resource.BinarySI)   // 128Mi
var memory256 = resource.NewQuantity(256*1024*1024, resource.BinarySI)   // 256Mi
var memory512 = resource.NewQuantity(512*1024*1024, resource.BinarySI)   // 512Mi
var memory2560 = resource.NewQuantity(2560*1024*1024, resource.BinarySI) // 2560Mi

var secretCheckCmd = `set -- $SECRET_LIST; ` +
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

var CommonEnvVars = []corev1.EnvVar{
	{
		Name:  "NODE_TLS_REJECT_UNAUTHORIZED",
		Value: "0",
	},
	{
		Name: "ICP_API_KEY",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "icp-serviceid-apikey-secret",
				},
				Key:      "ICP_API_KEY",
				Optional: &trueVar,
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

var commonSecretCheckVolumeMounts = []corev1.VolumeMount{
	{
		Name:      "icp-serviceid-apikey-secret",
		MountPath: "/sec/icp-serviceid-apikey-secret",
	},
	{
		Name:      "mongodb-ca-cert",
		MountPath: "/sec/cluster-ca-cert",
	},
	{
		Name:      "mongodb-client-cert",
		MountPath: "/sec/icp-mongodb-client-cert",
	},
	{
		Name:      "muser-icp-mongodb-admin",
		MountPath: "/sec/muser-icp-mongodb-admin",
	},
	{
		Name:      "mpass-icp-mongodb-admin",
		MountPath: "/sec/mpass-icp-mongodb-admin",
	},
}

var receiverCertVolumeMount = corev1.VolumeMount{
	Name:      "icp-metering-receiver-certs",
	MountPath: "/sec/icp-metering-receiver-secret",
}
var apiCertVolumeMount = corev1.VolumeMount{
	Name:      "icp-metering-api-certs",
	MountPath: "/sec/icp-metering-api-secret",
}

var dmSecretCheckVolumeMounts = append(commonSecretCheckVolumeMounts, receiverCertVolumeMount)
var rdrSecretCheckVolumeMounts = append(commonSecretCheckVolumeMounts, apiCertVolumeMount)

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
	AllowPrivilegeEscalation: &falseVar,
	Privileged:               &falseVar,
	ReadOnlyRootFilesystem:   &trueVar,
	RunAsNonRoot:             &trueVar,
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
	{
		Name:      "loglevel",
		MountPath: "/etc/config",
	},
}

var CommonIngressAnnotations = map[string]string{
	"app.kubernetes.io/managed-by": "operator",
	"kubernetes.io/ingress.class":  "ibm-icp-management",
}

var DmSecretCheckContainer = corev1.Container{
	Image:           "metering-data-manager",
	Name:            "metering-dm-secret-check",
	ImagePullPolicy: corev1.PullAlways,
	Command: []string{
		"sh",
		"-c",
		secretCheckCmd,
	},
	Env: []corev1.EnvVar{
		{
			Name:  "SECRET_LIST",
			Value: "icp-serviceid-apikey-secret icp-mongodb-admin icp-mongodb-admin cluster-ca-cert icp-mongodb-client-cert",
			//CS??? use instance.Spec.MongoDB.UsernameSecret/PasswordSecret for icp-mongodb-admin
			//CS??? use instance.Spec.MongoDB.ClientCertsSecret for icp-mongodb-client-cert
			//CS??? use instance.Spec.MongoDB.ClusterCertsSecret for cluster-ca-cert
			//CS??? Value: "icp-serviceid-apikey-secret icp-metering-receiver-secret icp-mongodb-admin icp-mongodb-admin
			//CS???         cluster-ca-cert icp-mongodb-client-cert",
		},
		{
			Name:  "SECRET_DIR_LIST",
			Value: "icp-serviceid-apikey-secret muser-icp-mongodb-admin mpass-icp-mongodb-admin cluster-ca-cert icp-mongodb-client-cert",
			//CS??? use instance.Spec.MongoDB.ClientCertsSecret for icp-mongodb-client-cert
			//CS??? use instance.Spec.MongoDB.ClusterCertsSecret for cluster-ca-cert
			//CS??? Value: "icp-serviceid-apikey-secret icp-metering-receiver-secret muser-icp-mongodb-admin mpass-icp-mongodb-admin
			//CS???         cluster-ca-cert icp-mongodb-client-cert",
		},
	},
	VolumeMounts:    dmSecretCheckVolumeMounts,
	Resources:       commonInitResources,
	SecurityContext: &commonSecurityContext,
}

var DmInitContainer = corev1.Container{
	Image:           "metering-data-manager",
	Name:            "metering-dm-init",
	ImagePullPolicy: corev1.PullAlways,
	Command: []string{
		"node",
		"/datamanager/lib/metering_init.js",
		"verifyOnlyMongo",
	},
	// CommonEnvVars and mongoDBEnvVars will be added by the controller
	Env: []corev1.EnvVar{
		{
			Name:  "MCM_VERBOSE",
			Value: "true",
		},
	},
	VolumeMounts:    commonInitVolumeMounts,
	Resources:       commonInitResources,
	SecurityContext: &commonSecurityContext,
}

var DmMainContainer = corev1.Container{
	Image: "metering-data-manager",
	//CS??? Image: "hyc-cloud-private-edge-docker-local.artifactory.swg-devops.com/ibmcom-amd64/metering-data-manager:3.3.1",
	Name:            "metering-dm",
	ImagePullPolicy: corev1.PullAlways,
	VolumeMounts: []corev1.VolumeMount{
		{
			Name:      "icp-metering-receiver-certs",
			MountPath: "/certs/metering-receiver",
		},
	},
	// CommonEnvVars and mongoDBEnvVars will be added by the controller
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
			Name:  "HC_DM_MCM_RECEIVER_ENABLED",
			Value: "false",
			//CS??? Value: "true",
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
			Name:  "HC_RECEIVER_SSL_CA",
			Value: "/certs/metering-receiver/ca.crt",
		},
		{
			Name:  "HC_RECEIVER_SSL_CERT",
			Value: "/certs/metering-receiver/tls.crt",
		},
		{
			Name:  "HC_RECEIVER_SSL_KEY",
			Value: "/certs/metering-receiver/tls.key",
		},
		{
			Name:  "DEFAULT_IAM_TOKEN_SERVICE_PORT",
			Value: "10443",
		},
		{
			Name:  "DEFAULT_IAM_PAP_SERVICE_PORT",
			Value: "39001",
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
				Scheme: "",
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
				Scheme: "",
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

var RdrSecretCheckContainer = corev1.Container{
	Image:           "metering-data-manager",
	Name:            "metering-reader-secret-check",
	ImagePullPolicy: corev1.PullAlways,
	Command: []string{
		"sh",
		"-c",
		secretCheckCmd,
	},
	Env: []corev1.EnvVar{
		{
			Name:  "SECRET_LIST",
			Value: "icp-metering-api-secret icp-serviceid-apikey-secret icp-mongodb-admin icp-mongodb-admin cluster-ca-cert icp-mongodb-client-cert",
		},
		{
			Name:  "SECRET_DIR_LIST",
			Value: "icp-metering-api-secret icp-serviceid-apikey-secret muser-icp-mongodb-admin mpass-icp-mongodb-admin cluster-ca-cert icp-mongodb-client-cert",
		},
	},
	VolumeMounts:    rdrSecretCheckVolumeMounts,
	Resources:       commonInitResources,
	SecurityContext: &commonSecurityContext,
}

var RdrInitContainer = corev1.Container{
	Image:           "metering-data-manager",
	Name:            "metering-reader-init",
	ImagePullPolicy: corev1.PullAlways,
	Command: []string{
		"node",
		"/datamanager/lib/metering_init.js",
		"verifyOnlyMongo",
	},
	// CommonEnvVars and mongoDBEnvVars will be added by the controller
	Env:             []corev1.EnvVar{},
	VolumeMounts:    commonInitVolumeMounts,
	Resources:       commonInitResources,
	SecurityContext: &commonSecurityContext,
}

var RdrMainContainer = corev1.Container{
	Image: "metering-data-manager",
	//CS??? Image: "hyc-cloud-private-edge-docker-local.artifactory.swg-devops.com/ibmcom-amd64/metering-data-manager:3.3.1",
	Name:            "metering-reader",
	ImagePullPolicy: corev1.PullAlways,
	VolumeMounts: []corev1.VolumeMount{
		{
			Name:      "icp-metering-api-certs",
			MountPath: "/certs/metering-api",
		},
	},
	// CommonEnvVars and mongoDBEnvVars will be added by the controller
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
					FieldPath: "spec.nodeName",
				},
			},
		},
		{
			Name:  "DEFAULT_IAM_TOKEN_SERVICE_PORT",
			Value: "10443",
		},
		{
			Name:  "DEFAULT_IAM_PAP_SERVICE_PORT",
			Value: "39001",
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
				Scheme: "",
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
				Scheme: "",
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
