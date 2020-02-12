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

package meteringui

import (
	"context"
	"reflect"
	gorun "runtime"
	"time"

	operatorv1alpha1 "github.com/ibm/ibm-metering-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-metering-operator/pkg/resources"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const meteringUICrType = "meteringui_cr"

var commonVolumes = []corev1.Volume{}

var mongoDBEnvVars = []corev1.EnvVar{}
var clusterEnvVars = []corev1.EnvVar{}

var log = logf.Log.WithName("controller_meteringui")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new MeteringUI Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMeteringUI{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("meteringui-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource MeteringUI
	err = c.Watch(&source.Kind{Type: &operatorv1alpha1.MeteringUI{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource "Deployment" and requeue the owner Metering
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.MeteringUI{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource "Service" and requeue the owner Metering
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.MeteringUI{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource "Ingress" and requeue the owner Metering
	err = c.Watch(&source.Kind{Type: &netv1.Ingress{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.MeteringUI{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileMeteringUI implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileMeteringUI{}

// ReconcileMeteringUI reconciles a MeteringUI object
type ReconcileMeteringUI struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a MeteringUI object and makes changes based on the state read
// and what is in the MeteringUI.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a UI Deployment and Service for each MeteringUI CR
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMeteringUI) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling MeteringUI")

	// if we need to create several resources, set a flag so we just requeue one time instead of after each create.
	needToRequeue := false

	// Fetch the MeteringUI CR instance
	instance := &operatorv1alpha1.MeteringUI{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("MeteringUI resource not found. Ignoring since object must be deleted")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get MeteringUI CR")
		return reconcile.Result{}, err
	}

	version := instance.Spec.Version
	reqLogger.Info("got MeteringUI instance, version=" + version)
	reqLogger.Info("Checking UI Service")
	// Check if the UI Service already exists, if not create a new one
	newService, err := r.serviceForUI(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	currentService := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.UIDeploymentName, Namespace: instance.Namespace}, currentService)
	if err != nil && errors.IsNotFound(err) {
		// Create a new Service
		reqLogger.Info("Creating a new UI Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
		err = r.client.Create(context.TODO(), newService)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue
			reqLogger.Info("UI Service already exists")
			needToRequeue = true
		} else if err != nil {
			reqLogger.Error(err, "Failed to create new UI Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
			return reconcile.Result{}, err
		} else {
			// Service created successfully - return and requeue
			needToRequeue = true
		}
	} else if err != nil {
		reqLogger.Error(err, "Failed to get UI Service")
		return reconcile.Result{}, err
	} else {
		// Found service, so send an update to k8s and let it determine if the resource has changed
		reqLogger.Info("Updating UI Service")
		// Can't copy the entire Spec because ClusterIP is immutable
		currentService.Spec.Ports = newService.Spec.Ports
		currentService.Spec.Selector = newService.Spec.Selector
		err = r.client.Update(context.TODO(), currentService)
		if err != nil {
			reqLogger.Error(err, "Failed to update UI Service", "Deployment.Namespace", currentService.Namespace,
				"Deployment.Name", currentService.Name)
			return reconcile.Result{}, err
		}
	}

	// set common MongoDB env vars based on the instance
	mongoDBEnvVars = res.BuildMongoDBEnvVars(instance.Spec.MongoDB.Host, instance.Spec.MongoDB.Port,
		instance.Spec.MongoDB.UsernameSecret, instance.Spec.MongoDB.UsernameKey,
		instance.Spec.MongoDB.PasswordSecret, instance.Spec.MongoDB.PasswordKey)
	// set common cluster env vars based on the instance
	clusterEnvVars = res.BuildUIClusterEnvVars(instance.Namespace, instance.Spec.IAMnamespace, instance.Spec.IngressNamespace,
		instance.Spec.External.ClusterName, instance.Spec.External.ClusterIP, instance.Spec.External.ClusterPort, false)

	// set common Volumes based on the instance
	commonVolumes = res.BuildCommonVolumes(instance.Spec.MongoDB.ClusterCertsSecret, instance.Spec.MongoDB.ClientCertsSecret,
		instance.Spec.MongoDB.UsernameSecret, instance.Spec.MongoDB.PasswordSecret, res.UIDeploymentName, "loglevel")

	reqLogger.Info("Checking UI Deployment")
	// Check if the UI Deployment already exists, if not create a new one
	newDeployment, err := r.deploymentForUI(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	currentDeployment := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.UIDeploymentName, Namespace: instance.Namespace}, currentDeployment)
	if err != nil && errors.IsNotFound(err) {
		// Create a new deployment
		reqLogger.Info("Creating a new UI Deployment", "Deployment.Namespace", newDeployment.Namespace, "Deployment.Name", newDeployment.Name)
		err = r.client.Create(context.TODO(), newDeployment)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue
			reqLogger.Info("UI Deployment already exists")
			needToRequeue = true
		} else if err != nil {
			reqLogger.Error(err, "Failed to create new UI Deployment", "Deployment.Namespace", newDeployment.Namespace,
				"Deployment.Name", newDeployment.Name)
			return reconcile.Result{}, err
		} else {
			// Deployment created successfully - return and requeue
			needToRequeue = true
		}
	} else if err != nil {
		reqLogger.Error(err, "Failed to get UI Deployment")
		return reconcile.Result{}, err
	} else {
		// Found deployment, so send an update to k8s and let it determine if the resource has changed
		reqLogger.Info("Updating UI Deployment")
		currentDeployment.Spec = newDeployment.Spec
		err = r.client.Update(context.TODO(), currentDeployment)
		if err != nil {
			reqLogger.Error(err, "Failed to update UI Deployment", "Deployment.Namespace", currentDeployment.Namespace,
				"Deployment.Name", currentDeployment.Name)
			return reconcile.Result{}, err
		}
	}

	reqLogger.Info("Checking UI Ingress")
	// Check if the Ingress already exists, if not create a new one
	err = r.reconcileIngress(instance, &needToRequeue)
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

	reqLogger.Info("Updating MeteringUI status")
	// Update the MeteringUI status with the pod names.
	// List the pods for this instance's deployment.
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Namespace),
		client.MatchingLabels(res.LabelsForSelector(res.UIDeploymentName, meteringUICrType, instance.Name)),
	}
	if err = r.client.List(context.TODO(), podList, listOpts...); err != nil {
		reqLogger.Error(err, "Failed to list pods", "MeteringUI.Namespace", instance.Namespace, "MeteringUI.Name", res.UIDeploymentName)
		return reconcile.Result{}, err
	}
	podNames := res.GetPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, instance.Status.Nodes) {
		instance.Status.Nodes = podNames
		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update MeteringUI status")
			return reconcile.Result{}, err
		}
	}

	reqLogger.Info("Reconciliation completed")
	// since we updated the status in the CR, sleep 5 seconds to allow the CR to be refreshed.
	time.Sleep(5 * time.Second)
	return reconcile.Result{}, nil
}

// deploymentForUI returns a UI Deployment object
func (r *ReconcileMeteringUI) deploymentForUI(instance *operatorv1alpha1.MeteringUI) (*appsv1.Deployment, error) {
	reqLogger := log.WithValues("func", "deploymentForUI", "instance.Name", instance.Name)
	metaLabels := res.LabelsForMetadata(res.UIDeploymentName)
	selectorLabels := res.LabelsForSelector(res.UIDeploymentName, meteringUICrType, instance.Name)
	podLabels := res.LabelsForPodMetadata(res.UIDeploymentName, meteringUICrType, instance.Name)

	var dmImage, uiImage, imageRegistry string
	if instance.Spec.ImageRegistry == "" {
		imageRegistry = res.DefaultImageRegistry
		reqLogger.Info("use default imageRegistry=" + imageRegistry)
	} else {
		imageRegistry = instance.Spec.ImageRegistry
		reqLogger.Info("use instance imageRegistry=" + imageRegistry)
	}
	dmImage = imageRegistry + "/" + res.DefaultDmImageName + ":" + res.DefaultDmImageTag + instance.Spec.ImageTagPostfix
	reqLogger.Info("dmImage=" + dmImage)
	uiImage = imageRegistry + "/" + res.DefaultUIImageName + ":" + res.DefaultUIImageTag + instance.Spec.ImageTagPostfix
	reqLogger.Info("uiImage=" + uiImage)

	// set the SECRET_LIST env var
	nameList := res.APIKeySecretName + " " + res.PlatformOidcSecretName + " " + res.CommonSecretCheckNames
	// set the SECRET_DIR_LIST env var
	dirList := res.APIKeySecretName + " " + res.PlatformOidcSecretName + " " + res.CommonSecretCheckDirs
	volumeMounts := append(res.CommonSecretCheckVolumeMounts, res.PlatformOidcVolumeMount, res.APIKeyVolumeMount)
	uiSecretCheckContainer := res.BuildSecretCheckContainer(res.UIDeploymentName, dmImage,
		res.SecretCheckCmd, nameList, dirList, volumeMounts)

	initEnvVars := []corev1.EnvVar{}
	initEnvVars = append(initEnvVars, res.CommonEnvVars...)
	initEnvVars = append(initEnvVars, mongoDBEnvVars...)
	uiInitContainer := res.BuildInitContainer(res.UIDeploymentName, dmImage, initEnvVars)

	uiMainContainer := res.UIMainContainer
	uiMainContainer.Image = uiImage
	uiMainContainer.Name = res.UIDeploymentName
	uiMainContainer.Env = append(uiMainContainer.Env, res.IAMEnvVars...)
	uiMainContainer.Env = append(uiMainContainer.Env, res.UIEnvVars...)
	uiMainContainer.Env = append(uiMainContainer.Env, clusterEnvVars...)
	uiMainContainer.Env = append(uiMainContainer.Env, res.CommonEnvVars...)
	uiMainContainer.Env = append(uiMainContainer.Env, mongoDBEnvVars...)
	uiMainContainer.VolumeMounts = append(uiMainContainer.VolumeMounts, res.CommonMainVolumeMounts...)

	uiVolumes := append(commonVolumes, res.APIKeyVolume, res.PlatformOidcVolume)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.UIDeploymentName,
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
					Labels: podLabels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName:            res.GetServiceAccountName(),
					NodeSelector:                  res.ManagementNodeSelector,
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
					Volumes: uiVolumes,
					InitContainers: []corev1.Container{
						uiSecretCheckContainer,
						uiInitContainer,
					},
					Containers: []corev1.Container{
						uiMainContainer,
					},
				},
			},
		},
	}
	// Set MeteringUI instance as the owner and controller of the Deployment
	err := controllerutil.SetControllerReference(instance, deployment, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for UI Deployment")
		return nil, err
	}
	return deployment, nil
}

// serviceForUI returns a UI Service object
func (r *ReconcileMeteringUI) serviceForUI(instance *operatorv1alpha1.MeteringUI) (*corev1.Service, error) {
	reqLogger := log.WithValues("func", "serviceForUI", "instance.Name", instance.Name)
	metaLabels := res.LabelsForMetadata(res.UIDeploymentName)
	selectorLabels := res.LabelsForSelector(res.UIDeploymentName, meteringUICrType, instance.Name)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.UIServiceName,
			Namespace: instance.Namespace,
			Labels:    metaLabels,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name: "dashboard",
					Port: 3130,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 3130,
					},
				},
			},
			Selector: selectorLabels,
		},
	}

	// Set MeteringUI instance as the owner and controller of the Service
	err := controllerutil.SetControllerReference(instance, service, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for UI Service")
		return nil, err
	}
	return service, nil
}

// Check if the Ingress already exists, if not create a new one.
// This function was created to reduce the cyclomatic complexity :)
func (r *ReconcileMeteringUI) reconcileIngress(instance *operatorv1alpha1.MeteringUI, needToRequeue *bool) error {
	reqLogger := log.WithValues("func", "reconcileIngress", "instance.Name", instance.Name)

	newIngress := res.BuildIngress(instance.Namespace, res.UIIngressData)
	// Set MeteringUI instance as the owner and controller of the Ingress
	err := controllerutil.SetControllerReference(instance, newIngress, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for UI Ingress", "Ingress.Namespace", newIngress.Namespace,
			"Ingress.Name", newIngress.Name)
		return err
	}
	currentIngress := &netv1.Ingress{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.UIIngressData.Name, Namespace: instance.Namespace}, currentIngress)
	if err != nil && errors.IsNotFound(err) {
		// Create a new Ingress
		reqLogger.Info("Creating a new UI Ingress", "Ingress.Namespace", newIngress.Namespace, "Ingress.Name", newIngress.Name)
		err = r.client.Create(context.TODO(), newIngress)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue
			reqLogger.Info("UI Ingress already exists")
			*needToRequeue = true
		} else if err != nil {
			reqLogger.Error(err, "Failed to create new UI Ingress", "Ingress.Namespace", newIngress.Namespace,
				"Ingress.Name", newIngress.Name)
			return err
		} else {
			// Ingress created successfully - return and requeue
			*needToRequeue = true
		}
	} else if err != nil {
		reqLogger.Error(err, "Failed to get UI Ingress")
		return err
	} else {
		// Found Ingress, so send an update to k8s and let it determine if the resource has changed
		reqLogger.Info("Updating UI Ingress", "Ingress.Name", newIngress.Name)
		currentIngress.Spec = newIngress.Spec
		err = r.client.Update(context.TODO(), currentIngress)
		if err != nil {
			reqLogger.Error(err, "Failed to update UI Ingress", "Ingress.Namespace", newIngress.Namespace,
				"Ingress.Name", newIngress.Name)
			return err
		}
	}
	return nil
}
