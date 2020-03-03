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

package meteringsender

import (
	"context"
	"reflect"
	gorun "runtime"
	"time"

	operatorv1alpha1 "github.com/ibm/ibm-metering-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-metering-operator/pkg/resources"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const meteringSenderCrType = "meteringsender_cr"

var commonVolumes = []corev1.Volume{}

var mongoDBEnvVars = []corev1.EnvVar{}
var clusterEnvVars = []corev1.EnvVar{}

var log = logf.Log.WithName("controller_meteringsender")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new MeteringSender Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMeteringSender{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("meteringsender-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource MeteringSender
	err = c.Watch(&source.Kind{Type: &operatorv1alpha1.MeteringSender{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource "Deployment" and requeue the owner MeteringSender
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.MeteringSender{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileMeteringSender implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileMeteringSender{}

// ReconcileMeteringSender reconciles a MeteringSender object
type ReconcileMeteringSender struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a MeteringSender object and makes changes based on the state read
// and what is in the MeteringSender.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Sender Deployment for each MeteringSender CR
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMeteringSender) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Name", request.Name)
	reqLogger.Info("Reconciling MeteringSender", "Request.Namespace", request.Namespace)

	// if we need to create several resources, set a flag so we just requeue one time instead of after each create.
	needToRequeue := false

	// Fetch the MeteringSender CR instance
	instance := &operatorv1alpha1.MeteringSender{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("MeteringSender resource not found. Ignoring since object must be deleted")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get MeteringSender CR")
		return reconcile.Result{}, err
	}

	version := instance.Spec.Version
	reqLogger.Info("got MeteringSender instance, version=" + version)
	reqLogger.Info("Checking Sender Deployment", "Deployment.Name", res.SenderDeploymentName)

	// set common MongoDB env vars based on the instance
	mongoDBEnvVars = res.BuildMongoDBEnvVars(instance.Spec.MongoDB)
	// set common cluster env vars based on the instance
	clusterEnvVars = res.BuildSenderClusterEnvVars(instance.Namespace, instance.Spec.Sender.ClusterNamespace,
		instance.Spec.Sender.ClusterName, instance.Spec.Sender.HubKubeConfigSecret)

	// set common Volumes based on the instance
	commonVolumes = res.BuildCommonVolumes(instance.Spec.MongoDB, res.SenderDeploymentName, "loglevel")

	// Check if the Sender Deployment already exists, if not create a new one
	newDeployment, err := r.deploymentForSender(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = res.ReconcileDeployment(r.client, instance.Namespace, res.SenderDeploymentName, "Sender", newDeployment, &needToRequeue)
	if err != nil {
		return reconcile.Result{}, err
	}

	if needToRequeue {
		// one or more resources was created, so requeue the request after 5 seconds
		reqLogger.Info("Requeue the request")
		// tried RequeueAfter but it is ignored because we're watching secondary resources.
		// so sleep instead to allow resources to be created by k8s.
		time.Sleep(5 * time.Second)
		return reconcile.Result{Requeue: true}, nil
	}

	reqLogger.Info("Updating MeteringSender status")
	// Update the MeteringSender status with the pod names.
	// List the pods for this instance's deployment.
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Namespace),
		client.MatchingLabels(res.LabelsForSelector(res.SenderDeploymentName, meteringSenderCrType, instance.Name)),
	}
	if err = r.client.List(context.TODO(), podList, listOpts...); err != nil {
		reqLogger.Error(err, "Failed to list pods", "MeteringSender.Namespace", instance.Namespace, "MeteringSender.Name", res.SenderDeploymentName)
		return reconcile.Result{}, err
	}
	podNames := res.GetPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, instance.Status.Nodes) {
		instance.Status.Nodes = podNames
		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update MeteringSender status")
			return reconcile.Result{}, err
		}
	}

	reqLogger.Info("Reconciliation completed")
	// since we updated the status in the CR, sleep 5 seconds to allow the CR to be refreshed.
	time.Sleep(5 * time.Second)
	return reconcile.Result{}, nil
}

// deploymentForSender returns a Sender Deployment object
func (r *ReconcileMeteringSender) deploymentForSender(instance *operatorv1alpha1.MeteringSender) (*appsv1.Deployment, error) {
	reqLogger := log.WithValues("func", "deploymentForSender", "instance.Name", instance.Name)
	metaLabels := res.LabelsForMetadata(res.SenderDeploymentName)
	selectorLabels := res.LabelsForSelector(res.SenderDeploymentName, meteringSenderCrType, instance.Name)
	podLabels := res.LabelsForPodMetadata(res.SenderDeploymentName, meteringSenderCrType, instance.Name)

	var senderImage, imageRegistry string
	if instance.Spec.ImageRegistry == "" {
		imageRegistry = res.DefaultImageRegistry
		reqLogger.Info("use default imageRegistry=" + imageRegistry)
	} else {
		imageRegistry = instance.Spec.ImageRegistry
		reqLogger.Info("use instance imageRegistry=" + imageRegistry)
	}
	senderImage = imageRegistry + "/" + res.DefaultSenderImageName + ":" + res.DefaultSenderImageTag + instance.Spec.ImageTagPostfix
	reqLogger.Info("senderImage=" + senderImage)

	senderSecretCheckContainer := res.BuildSecretCheckContainer(res.SenderDeploymentName, senderImage,
		res.SenderSecretCheckCmd, instance.Spec.MongoDB, nil)
	hubEnvVar := corev1.EnvVar{
		Name:  "HC_HUB_CONFIG",
		Value: instance.Spec.Sender.HubKubeConfigSecret,
	}
	senderSecretCheckContainer.Env = append(senderSecretCheckContainer.Env, hubEnvVar)

	initEnvVars := []corev1.EnvVar{
		{
			Name:  "MCM_VERBOSE",
			Value: "true",
		},
	}
	initEnvVars = append(initEnvVars, res.CommonEnvVars...)
	initEnvVars = append(initEnvVars, mongoDBEnvVars...)
	senderInitContainer := res.BuildInitContainer(res.SenderDeploymentName, senderImage, initEnvVars)

	senderMainContainer := res.SenderMainContainer
	senderMainContainer.Image = senderImage
	senderMainContainer.Name = res.SenderDeploymentName
	senderMainContainer.Env = append(senderMainContainer.Env, clusterEnvVars...)
	senderMainContainer.Env = append(senderMainContainer.Env, res.CommonEnvVars...)
	senderMainContainer.Env = append(senderMainContainer.Env, mongoDBEnvVars...)
	senderMainContainer.VolumeMounts = append(senderMainContainer.VolumeMounts, res.CommonMainVolumeMounts...)

	senderVolumes := commonVolumes

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.SenderDeploymentName,
			Namespace: instance.Namespace,
			Labels:    metaLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &res.Replica1,
			Selector: &metav1.LabelSelector{
				MatchLabels: selectorLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      podLabels,
					Annotations: res.AnnotationsForPod(),
				},
				Spec: corev1.PodSpec{
					ServiceAccountName:            res.GetServiceAccountName(),
					HostNetwork:                   false,
					HostPID:                       false,
					HostIPC:                       false,
					TerminationGracePeriodSeconds: &res.Seconds60,
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "beta.kubernetes.io/arch",
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{gorun.GOARCH},
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
					Volumes: senderVolumes,
					InitContainers: []corev1.Container{
						senderSecretCheckContainer,
						senderInitContainer,
					},
					Containers: []corev1.Container{
						senderMainContainer,
					},
				},
			},
		},
	}
	// Set MeteringSender instance as the owner and controller of the Deployment
	err := controllerutil.SetControllerReference(instance, deployment, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for Sender Deployment")
		return nil, err
	}
	return deployment, nil
}
