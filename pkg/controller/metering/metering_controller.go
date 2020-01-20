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

//CS??? use labelsForMeteringSelect to buld labelsForMeteringPod
//CS??? need to create icp-metering-receiver-secret; see metering-receiver-certificate.yaml
package metering

import (
	"context"
	"reflect"

	res "github.com/ibm/metering-operator/pkg/resources"

	operatorv1alpha1 "github.com/ibm/metering-operator/pkg/apis/operator/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const meteringComponentName = "meteringsvc"
const meteringReleaseName = "metering"
const dataManagerDeploymentName = "metering-dm"
const dataManagerImageTag = "3.3.1"

//const readerDaemonSetName = "metering-reader"

var trueVar = true
var defaultMode int32 = 420
var seconds60 int64 = 60
var nodeSelector = map[string]string{"management": "true"}

var commonVolumes = []corev1.Volume{}
var mongoDBEnvVars = []corev1.EnvVar{}

var log = logf.Log.WithName("controller_metering")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Metering Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMetering{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("metering-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Metering
	err = c.Watch(&source.Kind{Type: &operatorv1alpha1.Metering{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource "Deployment" and requeue the owner Metering
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.Metering{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource "Service" and requeue the owner Metering
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.Metering{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource "DaemonSet" and requeue the owner Metering
	err = c.Watch(&source.Kind{Type: &appsv1.DaemonSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.Metering{},
	})
	if err != nil {
		return err
	}
	return nil
}

// blank assignment to verify that ReconcileMetering implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileMetering{}

// ReconcileMetering reconciles a Metering object
type ReconcileMetering struct {
	// TODO: Clarify the split client
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Metering object and makes changes based on the state read
// and what is in the Metering.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a DataManager Deployment and Service for each Metering CR
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMetering) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Metering")

	// Fetch the Metering instance
	instance := &operatorv1alpha1.Metering{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("Metering resource not found. Ignoring since object must be deleted")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get Metering")
		return reconcile.Result{}, err
	}

	opVersion := instance.Spec.OperatorVersion
	reqLogger.Info("CS??? got Metering instance, version=" + opVersion + ", checking Service")
	// Check if the service already exists, if not create a new one
	currentService := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: dataManagerDeploymentName, Namespace: instance.Namespace}, currentService)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		newService := r.serviceForDataMgr(instance)
		reqLogger.Info("Creating a new Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
		err = r.client.Create(context.TODO(), newService)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
			return reconcile.Result{}, err
		}
		// Service created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Service")
		return reconcile.Result{}, err
	}

	reqLogger.Info("CS??? got Service, checking Deployment")
	// set common MongoDB env vars based on the instance
	mongoDBEnvVars = buildMongoEnvVars(instance)
	// set common Volumes based on the instance
	commonVolumes = buildCommonVolumes(instance)

	// Check if the deployment already exists, if not create a new one
	currentDeployment := &appsv1.Deployment{}
	//CS??? err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, currentDeployment)
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: dataManagerDeploymentName, Namespace: instance.Namespace}, currentDeployment)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		newDeployment := r.deploymentForDataMgr(instance)
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", newDeployment.Namespace, "Deployment.Name", newDeployment.Name)
		err = r.client.Create(context.TODO(), newDeployment)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", newDeployment.Namespace, "Deployment.Name", newDeployment.Name)
			return reconcile.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Deployment")
		return reconcile.Result{}, err
	}

	reqLogger.Info("CS??? got Deployment, checking DaemonSet")
	// Check if the DaemonSet already exists, if not create a new one
	/* CS???
	currentDaemonSet := &appsv1.DaemonSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: readerDaemonSetName, Namespace: instance.Namespace}, currentDaemonSet)
	if err != nil && errors.IsNotFound(err) {
		// Define a new DaemonSet
		newDaemonSet := r.daemonForReader(instance)
		reqLogger.Info("Creating a new DaemonSet", "Deployment.Namespace", newDaemonSet.Namespace, "Deployment.Name", newDaemonSet.Name)
		err = r.client.Create(context.TODO(), newDaemonSet)
		if err != nil {
			reqLogger.Error(err, "Failed to create new DaemonSet", "Deployment.Namespace", newDaemonSet.Namespace,
				"Deployment.Name", newDaemonSet.Name)
			return reconcile.Result{}, err
		}
		// DaemonSet created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get DaemonSet")
		return reconcile.Result{}, err
	}
	CS??? */

	reqLogger.Info("CS??? checking current deployment")
	// Ensure the deployment size is the same as the spec
	size := instance.Spec.DataManager.Size
	if *currentDeployment.Spec.Replicas != size {
		currentDeployment.Spec.Replicas = &size
		reqLogger.Info("CS??? updating current deployment")
		err = r.client.Update(context.TODO(), currentDeployment)
		if err != nil {
			reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace", currentDeployment.Namespace,
				"Deployment.Name", currentDeployment.Name)
			return reconcile.Result{}, err
		}
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}

	//CS??? reqLogger.Info("CS??? checking current DaemonSet")

	reqLogger.Info("CS??? updating Metering status")
	// Update the Metering status with the pod names
	// List the pods for this instance's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Namespace),
		client.MatchingLabels(labelsForMeteringSelect(instance.Name, dataManagerDeploymentName)),
	}
	if err = r.client.List(context.TODO(), podList, listOpts...); err != nil {
		reqLogger.Error(err, "Failed to list pods", "Metering.Namespace", instance.Namespace, "Metering.Name", dataManagerDeploymentName)
		return reconcile.Result{}, err
	}
	reqLogger.Info("CS??? get pod names")
	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, instance.Status.Nodes) {
		instance.Status.Nodes = podNames
		reqLogger.Info("CS??? put pod names in status")
		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update Metering status")
			return reconcile.Result{}, err
		}
	}

	reqLogger.Info("CS??? all done")
	return reconcile.Result{}, nil
}

// deploymentForDataMgr returns a DataManager Deployment object
func (r *ReconcileMetering) deploymentForDataMgr(instance *operatorv1alpha1.Metering) *appsv1.Deployment {
	reqLogger := log.WithValues("func", "deploymentForDataMgr", "instance.Name", instance.Name)
	labels1 := labelsForMeteringMeta(dataManagerDeploymentName)
	labels2 := labelsForMeteringSelect(instance.Name, dataManagerDeploymentName)
	labels3 := labelsForMeteringPod(instance.Name, dataManagerDeploymentName)

	replicas := instance.Spec.DataManager.Size
	image := instance.Spec.DataManager.ImageRegistry + ":" + dataManagerImageTag
	reqLogger.Info("CS??? image=" + image)

	res.DmSecretCheckContainer.Image = image
	res.DmSecretCheckContainer.Name = dataManagerDeploymentName + "-secret-check"

	res.DmInitContainer.Image = image
	res.DmInitContainer.Name = dataManagerDeploymentName + "-init"
	res.DmInitContainer.Env = append(res.DmInitContainer.Env, res.CommonEnvVars...)
	res.DmInitContainer.Env = append(res.DmInitContainer.Env, mongoDBEnvVars...)

	res.DmMainContainer.Image = image
	res.DmMainContainer.Name = dataManagerDeploymentName
	res.DmMainContainer.Env = append(res.DmMainContainer.Env, res.CommonEnvVars...)
	res.DmMainContainer.Env = append(res.DmMainContainer.Env, mongoDBEnvVars...)

	receiverCertVolume := corev1.Volume{
		Name: "icp-metering-receiver-certs",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName:  "icp-metering-receiver-secret",
				DefaultMode: &defaultMode,
				Optional:    &trueVar,
			},
		},
	}
	dmVolumes := append(commonVolumes, receiverCertVolume)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dataManagerDeploymentName,
			Namespace: instance.Namespace,
			Labels:    labels1,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels2,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels3,
				},
				Spec: corev1.PodSpec{
					NodeSelector:                  nodeSelector,
					PriorityClassName:             "system-cluster-critical",
					TerminationGracePeriodSeconds: &seconds60,
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "beta.kubernetes.io/arch",
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{"amd64"},
											},
										},
									},
								},
							},
						},
					},
					Tolerations: []corev1.Toleration{
						{
							Key:      "dedicated",
							Operator: corev1.TolerationOpExists,
							Effect:   corev1.TaintEffectNoSchedule,
						},
						{
							Key:      "CriticalAddonsOnly",
							Operator: corev1.TolerationOpExists,
						},
					},
					Volumes: dmVolumes,
					InitContainers: []corev1.Container{
						res.DmSecretCheckContainer,
						res.DmInitContainer,
					},
					Containers: []corev1.Container{
						res.DmMainContainer,
					},
				},
			},
		},
	}
	// Set Metering instance as the owner and controller of the Deployment
	err := controllerutil.SetControllerReference(instance, deployment, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for Deployment")
		return nil
	}
	return deployment
}

// serviceForDataMgr returns a DataManager Service object
func (r *ReconcileMetering) serviceForDataMgr(instance *operatorv1alpha1.Metering) *corev1.Service {
	reqLogger := log.WithValues("serviceForDataMgr", "Entry", "instance.Name", instance.Name)
	labels1 := labelsForMeteringMeta(dataManagerDeploymentName)
	labels2 := labelsForMeteringSelect(instance.Name, dataManagerDeploymentName)

	reqLogger.Info("CS??? Entry")
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        dataManagerDeploymentName,
			Namespace:   instance.Namespace,
			Labels:      labels1,
			Annotations: map[string]string{"prometheus.io/scrape": "false", "prometheus.io/scheme": "http"},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "datamanager",
					Port: 3000},
			},
			Selector: labels2,
		},
	}

	// Set Metering instance as the owner and controller of the Service
	err := controllerutil.SetControllerReference(instance, service, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for Service")
		return nil
	}
	return service
}

/*
// daemonForReader returns a Reader DaemonSet object
func (r *ReconcileMetering) daemonForReader(instance *operatorv1alpha1.Metering) *appsv1.DaemonSet {
	reqLogger := log.WithValues("serviceForDataMgr", "Entry", "instance.Name", instance.Name)
	labels1 := labelsForMeteringMeta(readerDaemonSetName)
	labels2 := labelsForMeteringSelect(instance.Name, readerDaemonSetName)
	labels3 := labelsForMeteringPod(instance.Name, readerDaemonSetName)

	reqLogger.Info("CS??? Entry")

	daemon := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      readerDaemonSetName,
			Namespace: instance.Namespace,
			Labels:    labels1,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels2,
			},
			UpdateStrategy: appsv1.DaemonSetUpdateStrategy{
				Type: appsv1.RollingUpdateDaemonSetStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDaemonSet{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 1,
					},
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels3,
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: &seconds60,
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "beta.kubernetes.io/arch",
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{"amd64"},
											},
										},
									},
								},
							},
						},
					},
					Tolerations: []corev1.Toleration{
						{
							Key:      "dedicated",
							Operator: corev1.TolerationOpExists,
							Effect:   corev1.TaintEffectNoSchedule,
						},
						{
							Key:      "CriticalAddonsOnly",
							Operator: corev1.TolerationOpExists,
						},
					},
					Volumes: commonVolumes,
					InitContainers: []corev1.Container{
						res.RdrSecretCheckContainer,
						res.RdrInitContainer,
					},
					Containers: []corev1.Container{
						res.RdrMainContainer,
					},
				},
			},
		},
	}

	// Set Metering instance as the owner and controller of the DaemonSet
	err := controllerutil.SetControllerReference(instance, daemon, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for DaemonSet")
		return nil
	}
	return daemon
}
*/

/*
// certForApi returns a Certificate object
func (r *ReconcileMetering) certForApi(instance *operatorv1alpha1.Metering) *certmgr.Certificate {
	reqLogger := log.WithValues("certForApi", "Entry", "instance.Name", instance.Name)
	reqLogger.Info("CS??? Entry")
	labels := labelsForMeteringSelect(instance.Name, readerDaemonSetName)

	certificate := &certmgr.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "icp-metering-api-ca-cert",
			Labels: labels,
		},
		Spec: certmgr.CertificateSpec{
			CommonName:   "metering-server",
			SecretName:   "icp-metering-api-secret",
			IsCA:         false,
			DNSNames:     []string{"metering-server"},
			Organization: []string{"IBM"},
			IssuerRef: certmgr.ObjectReference{
				Name: "icp-ca-issuer",
				Kind: certmgr.ClusterIssuerKind,
			},
		},
	}

	// Set Metering instance as the owner and controller of the Certificate
	err := controllerutil.SetControllerReference(instance, certificate, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for Certificate")
		return nil
	}
	return certificate
}
*/

func buildMongoEnvVars(instance *operatorv1alpha1.Metering) []corev1.EnvVar {
	mongoEnvVars := []corev1.EnvVar{
		{
			Name:  "HC_MONGO_HOST",
			Value: instance.Spec.MongoDB.Host,
		},
		{
			Name:  "HC_MONGO_PORT",
			Value: instance.Spec.MongoDB.Port,
		},
		{
			Name: "HC_MONGO_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: instance.Spec.MongoDB.UsernameSecret,
					},
					Key:      instance.Spec.MongoDB.UsernameKey,
					Optional: &trueVar,
				},
			},
		},
		{
			Name: "HC_MONGO_PASS",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: instance.Spec.MongoDB.PasswordSecret,
					},
					Key:      instance.Spec.MongoDB.PasswordKey,
					Optional: &trueVar,
				},
			},
		},
	}
	return mongoEnvVars
}

func buildCommonVolumes(instance *operatorv1alpha1.Metering) []corev1.Volume {
	commonVolumes := []corev1.Volume{
		{
			Name: "mongodb-ca-cert",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  instance.Spec.MongoDB.ClusterCertsSecret,
					DefaultMode: &defaultMode,
					Optional:    &trueVar,
				},
			},
		},
		{
			Name: "mongodb-client-cert",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  instance.Spec.MongoDB.ClientCertsSecret,
					DefaultMode: &defaultMode,
					Optional:    &trueVar,
				},
			},
		},
		{
			Name: "muser-icp-mongodb-admin",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  instance.Spec.MongoDB.UsernameSecret,
					DefaultMode: &defaultMode,
					Optional:    &trueVar,
				},
			},
		},
		{
			Name: "mpass-icp-mongodb-admin",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  instance.Spec.MongoDB.PasswordSecret,
					DefaultMode: &defaultMode,
					Optional:    &trueVar,
				},
			},
		},
		{
			Name: "icp-serviceid-apikey-secret",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  "icp-serviceid-apikey-secret",
					DefaultMode: &defaultMode,
					Optional:    &trueVar,
				},
			},
		},
	}
	return commonVolumes
}

// labelsForMetering returns the labels for selecting the resources
// belonging to the given metering CR name.
//CS??? need separate func for each image to set "instanceName"???
func labelsForMeteringPod(instanceName string, deploymentName string) map[string]string {
	return map[string]string{"app": deploymentName, "component": meteringComponentName, "metering_cr": instanceName,
		"app.kubernetes.io/name": deploymentName, "app.kubernetes.io/component": meteringComponentName, "release": meteringReleaseName}
	//CS??? return map[string]string{"app": deploymentName, "component": meteringComponentName, "metering_cr": instanceName}
	//CS??? return map[string]string{"app.kubernetes.io/name": deploymentName, "app.kubernetes.io/component": meteringComponentName,
	//CS??? "metering_cr": instanceName}
}

//CS??? need separate func for each image to set "app"???
func labelsForMeteringSelect(instanceName string, deploymentName string) map[string]string {
	return map[string]string{"app": deploymentName, "component": meteringComponentName, "metering_cr": instanceName}
}

//CS???
func labelsForMeteringMeta(deploymentName string) map[string]string {
	return map[string]string{"app.kubernetes.io/name": deploymentName, "app.kubernetes.io/component": meteringComponentName,
		"release": meteringReleaseName}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	reqLogger := log.WithValues("Request.Namespace", "CS??? namespace", "Request.Name", "CS???")
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
		reqLogger.Info("CS??? pod name=" + pod.Name)
	}
	return podNames
}
