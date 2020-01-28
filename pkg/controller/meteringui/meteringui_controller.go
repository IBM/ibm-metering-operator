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

const uiDeploymentName = "metering-ui"
const uiIngressName = "metering-ui"
const uiIngressPath = "/metering/"
const uiIngressPort int32 = 3130
const uiIngressServiceName = "metering-ui"
const meteringUICrType = "meteringui_cr"

var commonVolumes = []corev1.Volume{}

var mongoDBEnvVars = []corev1.EnvVar{}
var clusterEnvVars = []corev1.EnvVar{}

var uiIngressAnnotations = map[string]string{
	"icp.management.ibm.com/auth-type":      "id-token",
	"icp.management.ibm.com/rewrite-target": "/",
}

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

	reqLogger := log.WithValues("func", "add")
	reqLogger.Info("CS??? OS=" + gorun.GOOS + ", arch=" + gorun.GOARCH)

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
// a Pod as an example
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

	opVersion := instance.Spec.OperatorVersion
	reqLogger.Info("got MeteringUI instance, version=" + opVersion + ", checking UI Service")
	// Check if the UI Service already exists, if not create a new one
	currentService := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: uiDeploymentName, Namespace: instance.Namespace}, currentService)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Service
		newService := r.serviceForUI(instance)
		reqLogger.Info("Creating a new UI Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
		err = r.client.Create(context.TODO(), newService)
		if err != nil {
			reqLogger.Error(err, "Failed to create new UI Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
			return reconcile.Result{}, err
		}
		// Service created successfully - return and requeue
		needToRequeue = true
		//CS??? return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get UI Service")
		return reconcile.Result{}, err
	}

	reqLogger.Info("got UI Service, checking UI Deployment")
	// set common MongoDB env vars based on the instance
	mongoDBEnvVars = res.BuildMongoDBEnvVars(instance.Spec.MongoDB.Host, instance.Spec.MongoDB.Port,
		instance.Spec.MongoDB.UsernameSecret, instance.Spec.MongoDB.UsernameKey,
		instance.Spec.MongoDB.PasswordSecret, instance.Spec.MongoDB.PasswordKey)
	// set common cluster env vars based on the instance
	clusterEnvVars = buildClusterEnvVars(instance)
	// set common Volumes based on the instance
	commonVolumes = res.BuildCommonVolumes(instance.Spec.MongoDB.ClusterCertsSecret, instance.Spec.MongoDB.ClientCertsSecret,
		instance.Spec.MongoDB.UsernameSecret, instance.Spec.MongoDB.PasswordSecret, uiDeploymentName)

	// Check if the UI Deployment already exists, if not create a new one
	currentDeployment := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: uiDeploymentName, Namespace: instance.Namespace}, currentDeployment)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		newDeployment := r.deploymentForUI(instance)
		reqLogger.Info("Creating a new UI Deployment", "Deployment.Namespace", newDeployment.Namespace, "Deployment.Name", newDeployment.Name)
		err = r.client.Create(context.TODO(), newDeployment)
		if err != nil {
			reqLogger.Error(err, "Failed to create new UI Deployment", "Deployment.Namespace", newDeployment.Namespace,
				"Deployment.Name", newDeployment.Name)
			return reconcile.Result{}, err
		}
		// Deployment created successfully - return and requeue
		needToRequeue = true
		//CS??? return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get UI Deployment")
		return reconcile.Result{}, err
	}

	reqLogger.Info("got UI Deployment, checking Ingress")
	// Check if the Ingress already exists, if not create a new one
	currentIngress := &netv1.Ingress{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: uiIngressName, Namespace: instance.Namespace}, currentIngress)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Ingress
		newAnnotations := uiIngressAnnotations
		for key, value := range res.CommonIngressAnnotations {
			newAnnotations[key] = value
		}
		newIngress := res.BuildIngress(uiIngressName, instance.Namespace, uiIngressPath, uiIngressPort, newAnnotations, uiIngressServiceName)
		// Set MeteringUI instance as the owner and controller of the Ingress
		err = controllerutil.SetControllerReference(instance, newIngress, r.scheme)
		if err != nil {
			reqLogger.Error(err, "Failed to set owner for UI Ingress", "Ingress.Namespace", newIngress.Namespace,
				"Ingress.Name", newIngress.Name)
			return reconcile.Result{}, err
		}
		reqLogger.Info("Creating a new UI Ingress", "Ingress.Namespace", newIngress.Namespace, "Ingress.Name", newIngress.Name)
		err = r.client.Create(context.TODO(), newIngress)
		if err != nil {
			reqLogger.Error(err, "Failed to create new UI Ingress", "Ingress.Namespace", newIngress.Namespace,
				"Ingress.Name", newIngress.Name)
			return reconcile.Result{}, err
		}
		// Ingress created successfully - return and requeue
		needToRequeue = true
		//CS??? return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get UI Ingress")
		return reconcile.Result{}, err
	}

	if needToRequeue {
		// one or more resources was created, so requeue the request
		reqLogger.Info("Requeue the request")
		return reconcile.Result{Requeue: true}, nil
	}

	reqLogger.Info("got UI Ingress, checking current UI deployment")
	// Ensure the image is the same as the spec
	var expectedImage string
	if instance.Spec.ImageRegistry == "" {
		expectedImage = res.DefaultImageRegistry + "/" + res.DefaultUIImageName + ":" + res.DefaultUIImageTag
		reqLogger.Info("CS??? default expectedImage=" + expectedImage)
	} else {
		expectedImage = instance.Spec.ImageRegistry + "/" + res.DefaultUIImageName + ":" + res.DefaultUIImageTag
		reqLogger.Info("CS??? expectedImage=" + expectedImage)
	}
	if currentDeployment.Spec.Template.Spec.Containers[0].Image != expectedImage {
		reqLogger.Info("CS??? curr image=" + currentDeployment.Spec.Template.Spec.Containers[0].Image + ", expect=" + expectedImage)
		currentDeployment.Spec.Template.Spec.Containers[0].Image = expectedImage
		reqLogger.Info("updating current UI deployment")
		err = r.client.Update(context.TODO(), currentDeployment)
		if err != nil {
			reqLogger.Error(err, "Failed to update UI Deployment", "Deployment.Namespace", currentDeployment.Namespace,
				"Deployment.Name", currentDeployment.Name)
			return reconcile.Result{}, err
		}
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}

	reqLogger.Info("Updating MeteringUI status")
	// Update the MeteringUI status with the pod names
	// List the pods for this instance's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Namespace),
		client.MatchingLabels(res.LabelsForSelector(uiDeploymentName, meteringUICrType, instance.Name)),
	}
	if err = r.client.List(context.TODO(), podList, listOpts...); err != nil {
		reqLogger.Error(err, "Failed to list pods", "MeteringUI.Namespace", instance.Namespace, "MeteringUI.Name", uiDeploymentName)
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
			reqLogger.Error(err, "Failed to update MeteringUI status")
			return reconcile.Result{}, err
		}
	}

	reqLogger.Info("CS??? all done")
	return reconcile.Result{}, nil
}

// deploymentForUI returns a UI Deployment object
func (r *ReconcileMeteringUI) deploymentForUI(instance *operatorv1alpha1.MeteringUI) *appsv1.Deployment {
	reqLogger := log.WithValues("func", "deploymentForUI", "instance.Name", instance.Name)
	metaLabels := res.LabelsForMetadata(uiDeploymentName)
	selectorLabels := res.LabelsForSelector(uiDeploymentName, meteringUICrType, instance.Name)
	podLabels := res.LabelsForPodMetadata(uiDeploymentName, meteringUICrType, instance.Name)

	var dmImage string
	var uiImage string
	if instance.Spec.ImageRegistry == "" {
		dmImage = res.DefaultImageRegistry + "/" + res.DefaultDmImageName + ":" + res.DefaultDmImageTag
		reqLogger.Info("CS??? default dmImage=" + dmImage)
		uiImage = res.DefaultImageRegistry + "/" + res.DefaultUIImageName + ":" + res.DefaultUIImageTag
		reqLogger.Info("CS??? default uiImage=" + uiImage)
	} else {
		dmImage = instance.Spec.ImageRegistry + "/" + res.DefaultDmImageName + ":" + res.DefaultDmImageTag
		reqLogger.Info("CS??? dmImage=" + dmImage)
		uiImage = instance.Spec.ImageRegistry + "/" + res.DefaultUIImageName + ":" + res.DefaultUIImageTag
		reqLogger.Info("CS??? uiImage=" + uiImage)
	}

	uiSecretCheckContainer := res.UISecretCheckContainer
	uiSecretCheckContainer.Image = dmImage
	uiSecretCheckContainer.Name = uiDeploymentName + "-secret-check"

	uiInitContainer := res.UIInitContainer
	uiInitContainer.Image = dmImage
	uiInitContainer.Name = uiDeploymentName + "-init"
	uiInitContainer.Env = append(uiInitContainer.Env, res.CommonEnvVars...)
	uiInitContainer.Env = append(uiInitContainer.Env, mongoDBEnvVars...)

	uiMainContainer := res.UIMainContainer
	uiMainContainer.Image = uiImage
	uiMainContainer.Name = uiDeploymentName
	uiMainContainer.Env = append(uiMainContainer.Env, res.CommonEnvVars...)
	uiMainContainer.Env = append(uiMainContainer.Env, mongoDBEnvVars...)
	uiMainContainer.Env = append(uiMainContainer.Env, clusterEnvVars...)
	uiMainContainer.VolumeMounts = append(uiMainContainer.VolumeMounts, res.CommonMainVolumeMounts...)

	platformOidcVolume := corev1.Volume{
		Name: "platform-oidc-credentials",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName:  "platform-oidc-credentials",
				DefaultMode: &res.DefaultMode,
				Optional:    &res.TrueVar,
			},
		},
	}
	uiVolumes := append(commonVolumes, platformOidcVolume)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      uiDeploymentName,
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
		return nil
	}
	return deployment
}

// serviceForUI returns a UI Service object
func (r *ReconcileMeteringUI) serviceForUI(instance *operatorv1alpha1.MeteringUI) *corev1.Service {
	reqLogger := log.WithValues("func", "serviceForUI", "instance.Name", instance.Name)
	metaLabels := res.LabelsForMetadata(uiDeploymentName)
	selectorLabels := res.LabelsForSelector(uiDeploymentName, meteringUICrType, instance.Name)

	reqLogger.Info("CS??? Entry")
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      uiDeploymentName,
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
		return nil
	}
	return service
}

func buildClusterEnvVars(instance *operatorv1alpha1.MeteringUI) []corev1.EnvVar {
	reqLogger := log.WithValues("func", "buildClusterEnvVars", "instance.Name", instance.Name)
	reqLogger.Info("CS??? Entry")

	var iamNamespace string
	if instance.Spec.IAMnamespace != "" {
		reqLogger.Info("CS??? IAMnamespace=" + instance.Spec.IAMnamespace)
		iamNamespace = instance.Spec.IAMnamespace
	} else {
		reqLogger.Info("CS??? IAMnamespace is blank, use instance")
		iamNamespace = instance.Namespace
	}

	clusterEnvVars := []corev1.EnvVar{
		{
			Name:  "CLUSTER_NAME",
			Value: instance.Spec.External.ClusterName,
		},
		{
			Name:  "ICP_EXTERNAL_IP",
			Value: instance.Spec.External.ClusterIP,
		},
		{
			Name:  "ICP_EXTERNAL_PORT",
			Value: instance.Spec.External.ClusterPort,
		},
		{
			Name:  "cfcRouterUrl",
			Value: instance.Spec.External.CfcRouterURL,
		},
		{
			Name:  "IAM_NAMESPACE",
			Value: iamNamespace,
		},
	}
	return clusterEnvVars
}
