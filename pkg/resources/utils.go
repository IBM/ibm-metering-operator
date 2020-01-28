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
	netv1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const MeteringComponentName = "meteringsvc"
const MeteringReleaseName = "metering"

var DefaultMode int32 = 420

var log = logf.Log.WithName("resource_utils")

// BuildIngress returns an Ingress object.
// Call controllerutil.SetControllerReference to set the owner and controller
// for the Ingress object created by this function.
func BuildIngress(name string, namespace string, path string,
	port int32, annotations map[string]string, serviceName string) *netv1.Ingress {
	reqLogger := log.WithValues("func", "buildIngress", "Ingress.Name", name)
	reqLogger.Info("CS??? Entry")
	metaLabels := labelsForIngressMeta(name)

	ingress := &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: annotations,
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
									Path: path,
									Backend: netv1.IngressBackend{
										ServiceName: serviceName,
										ServicePort: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: port,
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
	}
	return mongoDBEnvVars
}

func BuildCommonVolumes(clusterSecret string, clientSecret string, usernameSecret string,
	passwordSecret string, loglevelPrefix string) []corev1.Volume {
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
			Name: "icp-serviceid-apikey-secret",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  "icp-serviceid-apikey-secret",
					DefaultMode: &DefaultMode,
					Optional:    &TrueVar,
				},
			},
		},
		{
			Name: "loglevel",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "metering-logging-configuration",
					},
					Items: []corev1.KeyToPath{
						{
							Key:  loglevelPrefix + "-loglevel.json",
							Path: "loglevel.json",
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
