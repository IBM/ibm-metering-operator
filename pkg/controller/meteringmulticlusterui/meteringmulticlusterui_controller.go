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

package meteringmulticlusterui

import (
	"context"
	"reflect"
	gorun "runtime"

	operatorv1alpha1 "github.com/ibm/metering-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/metering-operator/pkg/resources"

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

const meteringMcmUICrType = "meteringmulticlusterui_cr"

var commonVolumes = []corev1.Volume{}

var mongoDBEnvVars = []corev1.EnvVar{}
var clusterEnvVars = []corev1.EnvVar{}

var log = logf.Log.WithName("controller_meteringmcmui")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new MeteringMultiClusterUI Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMeteringMultiClusterUI{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("meteringmulticlusterui-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	reqLogger := log.WithValues("func", "add")
	reqLogger.Info("CS??? OS=" + gorun.GOOS + ", arch=" + gorun.GOARCH)

	// Watch for changes to primary resource MeteringMultiClusterUI
	err = c.Watch(&source.Kind{Type: &operatorv1alpha1.MeteringMultiClusterUI{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource "Deployment" and requeue the owner Metering
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.MeteringMultiClusterUI{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource "Service" and requeue the owner Metering
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.MeteringMultiClusterUI{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource "Ingress" and requeue the owner Metering
	err = c.Watch(&source.Kind{Type: &netv1.Ingress{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.MeteringMultiClusterUI{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileMeteringMultiClusterUI implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileMeteringMultiClusterUI{}

// ReconcileMeteringMultiClusterUI reconciles a MeteringMultiClusterUI object
type ReconcileMeteringMultiClusterUI struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a MeteringMultiClusterUI object and makes changes based on the state read
// and what is in the MeteringMultiClusterUI.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMeteringMultiClusterUI) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling MeteringMultiClusterUI")

	// if we need to create several resources, set a flag so we just requeue one time instead of after each create.
	needToRequeue := false

	// Fetch the MeteringMultiClusterUI CR instance
	instance := &operatorv1alpha1.MeteringMultiClusterUI{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("MeteringMultiClusterUI resource not found. Ignoring since object must be deleted")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get MeteringMultiClusterUI CR")
		return reconcile.Result{}, err
	}

	opVersion := instance.Spec.OperatorVersion
	reqLogger.Info("got MeteringMultiClusterUI instance, version=" + opVersion + ", checking MCM UI Service")
	// Check if the MCM UI Service already exists, if not create a new one
	currentService := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.McmDeploymentName, Namespace: instance.Namespace}, currentService)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Service
		newService := r.serviceForMCMUI(instance)
		reqLogger.Info("Creating a new MCM UI Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
		err = r.client.Create(context.TODO(), newService)
		if err != nil {
			reqLogger.Error(err, "Failed to create new MCM UI Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
			return reconcile.Result{}, err
		}
		// Service created successfully - return and requeue
		needToRequeue = true
	} else if err != nil {
		reqLogger.Error(err, "Failed to get MCM UI Service")
		return reconcile.Result{}, err
	}

	// set common MongoDB env vars based on the instance
	mongoDBEnvVars = res.BuildMongoDBEnvVars(instance.Spec.MongoDB.Host, instance.Spec.MongoDB.Port,
		instance.Spec.MongoDB.UsernameSecret, instance.Spec.MongoDB.UsernameKey,
		instance.Spec.MongoDB.PasswordSecret, instance.Spec.MongoDB.PasswordKey)
	// set common cluster env vars based on the instance
	clusterEnvVars = res.BuildUIClusterEnvVars(instance.Namespace, instance.Spec.IAMnamespace, instance.Spec.IngressNamespace,
		instance.Spec.External.ClusterName, instance.Spec.External.ClusterIP, instance.Spec.External.ClusterPort, true)

	// set common Volumes based on the instance
	commonVolumes = res.BuildCommonVolumes(instance.Spec.MongoDB.ClusterCertsSecret, instance.Spec.MongoDB.ClientCertsSecret,
		instance.Spec.MongoDB.UsernameSecret, instance.Spec.MongoDB.PasswordSecret, res.McmDeploymentName, "log4js")

	reqLogger.Info("got MCM UI Service, checking MCM UI Deployment")
	// Check if the MCM UI Deployment already exists, if not create a new one
	currentDeployment := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.McmDeploymentName, Namespace: instance.Namespace}, currentDeployment)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		newDeployment := r.deploymentForMCMUI(instance)
		reqLogger.Info("Creating a new MCM UI Deployment", "Deployment.Namespace", newDeployment.Namespace, "Deployment.Name", newDeployment.Name)
		err = r.client.Create(context.TODO(), newDeployment)
		if err != nil {
			reqLogger.Error(err, "Failed to create new MCM UI Deployment", "Deployment.Namespace", newDeployment.Namespace,
				"Deployment.Name", newDeployment.Name)
			return reconcile.Result{}, err
		}
		// Deployment created successfully - return and requeue
		needToRequeue = true
	} else if err != nil {
		reqLogger.Error(err, "Failed to get MCM UI Deployment")
		return reconcile.Result{}, err
	}

	reqLogger.Info("got MCM UI Deployment, checking Ingress")
	// Check if the Ingress already exists, if not create a new one
	currentIngress := &netv1.Ingress{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.McmIngressData.Name, Namespace: instance.Namespace}, currentIngress)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Ingress
		newIngress := res.BuildIngress(instance.Namespace, res.McmIngressData)
		// Set MeteringMultiClusterUI instance as the owner and controller of the Ingress
		err = controllerutil.SetControllerReference(instance, newIngress, r.scheme)
		if err != nil {
			reqLogger.Error(err, "Failed to set owner for MCM UI Ingress", "Ingress.Namespace", newIngress.Namespace,
				"Ingress.Name", newIngress.Name)
			return reconcile.Result{}, err
		}
		reqLogger.Info("Creating a new MCM UI Ingress", "Ingress.Namespace", newIngress.Namespace, "Ingress.Name", newIngress.Name)
		err = r.client.Create(context.TODO(), newIngress)
		if err != nil {
			reqLogger.Error(err, "Failed to create new MCM UI Ingress", "Ingress.Namespace", newIngress.Namespace,
				"Ingress.Name", newIngress.Name)
			return reconcile.Result{}, err
		}
		// Ingress created successfully - return and requeue
		needToRequeue = true
	} else if err != nil {
		reqLogger.Error(err, "Failed to get MCM UI Ingress")
		return reconcile.Result{}, err
	}

	if needToRequeue {
		// one or more resources was created, so requeue the request
		reqLogger.Info("Requeue the request")
		return reconcile.Result{Requeue: true}, nil
	}

	reqLogger.Info("got MCM UI Ingress, checking current MCM UI deployment")
	// Ensure the image is the same as the spec
	var expectedImage string
	if instance.Spec.ImageRegistry == "" {
		expectedImage = res.DefaultImageRegistry + "/" + res.DefaultMcmUIImageName + ":" + res.DefaultMcmUIImageTag
		reqLogger.Info("CS??? default expectedImage=" + expectedImage)
	} else {
		expectedImage = instance.Spec.ImageRegistry + "/" + res.DefaultMcmUIImageName + ":" + res.DefaultMcmUIImageTag
		reqLogger.Info("CS??? expectedImage=" + expectedImage)
	}
	if currentDeployment.Spec.Template.Spec.Containers[0].Image != expectedImage {
		reqLogger.Info("CS??? curr image=" + currentDeployment.Spec.Template.Spec.Containers[0].Image + ", expect=" + expectedImage)
		currentDeployment.Spec.Template.Spec.Containers[0].Image = expectedImage
		reqLogger.Info("updating current MCM UI deployment")
		err = r.client.Update(context.TODO(), currentDeployment)
		if err != nil {
			reqLogger.Error(err, "Failed to update MCM UI Deployment", "Deployment.Namespace", currentDeployment.Namespace,
				"Deployment.Name", currentDeployment.Name)
			return reconcile.Result{}, err
		}
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}

	reqLogger.Info("Updating MeteringMultiClusterUI status")
	// Update the MeteringMultiClusterUI status with the pod names
	// List the pods for this instance's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Namespace),
		client.MatchingLabels(res.LabelsForSelector(res.McmDeploymentName, meteringMcmUICrType, instance.Name)),
	}
	if err = r.client.List(context.TODO(), podList, listOpts...); err != nil {
		reqLogger.Error(err, "Failed to list pods", "MeteringMultiClusterUI.Namespace", instance.Namespace, "MeteringMultiClusterUI.Name", res.McmDeploymentName)
		return reconcile.Result{}, err
	}
	reqLogger.Info("CS??? get pod names")
	podNames := res.GetPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, instance.Status.Nodes) {
		instance.Status.Nodes = podNames
		reqLogger.Info("CS??? put pod names in status")
		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update MeteringMultiClusterUI status")
			return reconcile.Result{}, err
		}
	}

	reqLogger.Info("CS??? all done")
	return reconcile.Result{}, nil
}

// deploymentForMCMUI returns an MCM UI Deployment object
func (r *ReconcileMeteringMultiClusterUI) deploymentForMCMUI(instance *operatorv1alpha1.MeteringMultiClusterUI) *appsv1.Deployment {
	reqLogger := log.WithValues("func", "deploymentForMCMUI", "instance.Name", instance.Name)
	metaLabels := res.LabelsForMetadata(res.McmDeploymentName)
	selectorLabels := res.LabelsForSelector(res.McmDeploymentName, meteringMcmUICrType, instance.Name)
	podLabels := res.LabelsForPodMetadata(res.McmDeploymentName, meteringMcmUICrType, instance.Name)

	var dmImage string
	var mcmImage string
	if instance.Spec.ImageRegistry == "" {
		dmImage = res.DefaultImageRegistry + "/" + res.DefaultDmImageName + ":" + res.DefaultDmImageTag
		reqLogger.Info("CS??? default dmImage=" + dmImage)
		mcmImage = res.DefaultImageRegistry + "/" + res.DefaultMcmUIImageName + ":" + res.DefaultMcmUIImageTag
		reqLogger.Info("CS??? default mcmImage=" + mcmImage)
	} else {
		dmImage = instance.Spec.ImageRegistry + "/" + res.DefaultDmImageName + ":" + res.DefaultDmImageTag
		reqLogger.Info("CS??? dmImage=" + dmImage)
		mcmImage = instance.Spec.ImageRegistry + "/" + res.DefaultMcmUIImageName + ":" + res.DefaultMcmUIImageTag
		reqLogger.Info("CS??? mcmImage=" + mcmImage)
	}

	mcmSecretCheckContainer := res.BaseSecretCheckContainer
	mcmSecretCheckContainer.Image = dmImage
	mcmSecretCheckContainer.Name = res.McmDeploymentName + "-secret-check"
	// set the SECRET_LIST env var
	mcmSecretCheckContainer.Env[res.SecretListVarNdx].Value = res.APIKeyZecretName + " " +
		res.PlatformOidcZecretName + " " + res.CommonZecretCheckNames
	// set the SECRET_DIR_LIST env var
	mcmSecretCheckContainer.Env[res.SecretDirVarNdx].Value = res.APIKeyZecretName + " " +
		res.PlatformOidcZecretName + " " + res.CommonZecretCheckDirs
	mcmSecretCheckContainer.VolumeMounts = append(res.CommonSecretCheckVolumeMounts, res.PlatformOidcVolumeMount, res.APIKeyVolumeMount)

	mcmInitContainer := res.BaseInitContainer
	mcmInitContainer.Image = dmImage
	mcmInitContainer.Name = res.McmDeploymentName + "-init"
	mcmInitContainer.Env = append(mcmInitContainer.Env, res.CommonEnvVars...)
	mcmInitContainer.Env = append(mcmInitContainer.Env, mongoDBEnvVars...)

	mcmMainContainer := res.McmUIMainContainer
	mcmMainContainer.Image = mcmImage
	mcmMainContainer.Name = res.McmDeploymentName
	mcmMainContainer.Env = append(mcmMainContainer.Env, res.CommonEnvVars...)
	mcmMainContainer.Env = append(mcmMainContainer.Env, res.IAMEnvVars...)
	mcmMainContainer.Env = append(mcmMainContainer.Env, res.UIEnvVars...)
	mcmMainContainer.Env = append(mcmMainContainer.Env, mongoDBEnvVars...)
	mcmMainContainer.Env = append(mcmMainContainer.Env, clusterEnvVars...)
	mcmMainContainer.VolumeMounts = append(mcmMainContainer.VolumeMounts, res.CommonMainVolumeMounts...)

	mcmVolumes := append(commonVolumes, res.APIKeyVolume, res.PlatformOidcVolume)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.McmDeploymentName,
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
					Volumes: mcmVolumes,
					InitContainers: []corev1.Container{
						mcmSecretCheckContainer,
						mcmInitContainer,
					},
					Containers: []corev1.Container{
						mcmMainContainer,
					},
				},
			},
		},
	}
	// Set MeteringMultiClusterUI instance as the owner and controller of the Deployment
	err := controllerutil.SetControllerReference(instance, deployment, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for MCM UI Deployment")
		return nil
	}
	return deployment
}

// serviceForMCMUI returns an MCM UI Service object
func (r *ReconcileMeteringMultiClusterUI) serviceForMCMUI(instance *operatorv1alpha1.MeteringMultiClusterUI) *corev1.Service {
	reqLogger := log.WithValues("func", "serviceForMCMUI", "instance.Name", instance.Name)
	metaLabels := res.LabelsForMetadata(res.McmDeploymentName)
	selectorLabels := res.LabelsForSelector(res.McmDeploymentName, meteringMcmUICrType, instance.Name)

	reqLogger.Info("CS??? Entry")
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.McmDeploymentName,
			Namespace: instance.Namespace,
			Labels:    metaLabels,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name: "metering-mcm-dashboard",
					Port: 3001,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 3001,
					},
				},
			},
			Selector: selectorLabels,
		},
	}

	// Set MeteringMultiClusterUI instance as the owner and controller of the Service
	err := controllerutil.SetControllerReference(instance, service, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for MCM UI Service")
		return nil
	}
	return service
}
