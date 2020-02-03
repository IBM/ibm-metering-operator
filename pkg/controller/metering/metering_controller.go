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

package metering

import (
	"context"
	"reflect"
	gorun "runtime"

	res "github.com/ibm/metering-operator/pkg/resources"

	"k8s.io/apimachinery/pkg/util/intstr"

	certmgr "github.com/ibm/metering-operator/pkg/apis/certmanager/v1alpha1"
	operatorv1alpha1 "github.com/ibm/metering-operator/pkg/apis/operator/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1beta1"
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

const meteringCrType = "metering_cr"

var commonVolumes = []corev1.Volume{}

var mongoDBEnvVars = []corev1.EnvVar{}
var clusterEnvVars = []corev1.EnvVar{}

var certificateList = []res.CertificateData{
	res.APICertificateData,
	res.ReceiverCertificateData,
}
var ingressList = []res.IngressData{
	res.APIcheckIngressData,
	res.APIrbacIngressData,
	res.APIswaggerIngressData,
}

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

	reqLogger := log.WithValues("func", "add")
	reqLogger.Info("CS??? OS=" + gorun.GOOS + ", arch=" + gorun.GOARCH)

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

	// Watch for changes to secondary resource "Ingress" and requeue the owner Metering
	err = c.Watch(&source.Kind{Type: &netv1.Ingress{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.Metering{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource "Certificate" and requeue the owner Metering
	/* CS???
	err = c.Watch(&source.Kind{Type: &certmgr.Certificate{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.Metering{},
	})
	if err != nil {
		return err
	}
	CS??? */

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

	// if we need to create several resources, set a flag so we just requeue one time instead of after each create.
	needToRequeue := false

	// Fetch the Metering CR instance
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
		reqLogger.Error(err, "Failed to get Metering CR")
		return reconcile.Result{}, err
	}

	opVersion := instance.Spec.OperatorVersion
	reqLogger.Info("got Metering instance, version=" + opVersion + ", checking Services")
	// Check if the DM and Reader Services already exist. If not, create a new one.
	err = r.reconcileService(instance, &needToRequeue)
	if err != nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("got Services, checking Certificates")
	// Check if the Certificates already exist, if not create new ones
	err = r.reconcileCertificate(instance, &needToRequeue)
	if err != nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("got Certificates, checking DM Deployment")
	// set common MongoDB env vars based on the instance
	mongoDBEnvVars = res.BuildMongoDBEnvVars(instance.Spec.MongoDB.Host, instance.Spec.MongoDB.Port,
		instance.Spec.MongoDB.UsernameSecret, instance.Spec.MongoDB.UsernameKey,
		instance.Spec.MongoDB.PasswordSecret, instance.Spec.MongoDB.PasswordKey)
	// set common cluster env vars based on the instance
	clusterEnvVars = res.BuildCommonClusterEnvVars(instance.Namespace, instance.Spec.IAMnamespace, instance.Spec.External.ClusterName)

	// set common Volumes based on the instance
	commonVolumes = res.BuildCommonVolumes(instance.Spec.MongoDB.ClusterCertsSecret, instance.Spec.MongoDB.ClientCertsSecret,
		instance.Spec.MongoDB.UsernameSecret, instance.Spec.MongoDB.PasswordSecret, res.DmDeploymentName, "loglevel")

	// Check if the DM Deployment already exists, if not create a new one
	currentDeployment := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.DmDeploymentName, Namespace: instance.Namespace}, currentDeployment)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		newDeployment := r.deploymentForDataMgr(instance)
		reqLogger.Info("Creating a new DM Deployment", "Deployment.Namespace", newDeployment.Namespace, "Deployment.Name", newDeployment.Name)
		err = r.client.Create(context.TODO(), newDeployment)
		if err != nil {
			reqLogger.Error(err, "Failed to create new DM Deployment", "Deployment.Namespace", newDeployment.Namespace,
				"Deployment.Name", newDeployment.Name)
			return reconcile.Result{}, err
		}
		// Deployment created successfully - return and requeue
		needToRequeue = true
	} else if err != nil {
		reqLogger.Error(err, "Failed to get DM Deployment")
		return reconcile.Result{}, err
	}

	reqLogger.Info("got DM Deployment, checking Rdr DaemonSet")
	// Check if the DaemonSet already exists, if not create a new one
	currentDaemonSet := &appsv1.DaemonSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.ReaderDaemonSetName, Namespace: instance.Namespace}, currentDaemonSet)
	if err != nil && errors.IsNotFound(err) {
		// Define a new DaemonSet
		newDaemonSet := r.daemonForReader(instance)
		reqLogger.Info("Creating a new Rdr DaemonSet", "DaemonSet.Namespace", newDaemonSet.Namespace, "DaemonSet.Name", newDaemonSet.Name)
		err = r.client.Create(context.TODO(), newDaemonSet)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Rdr DaemonSet", "DaemonSet.Namespace", newDaemonSet.Namespace,
				"DaemonSet.Name", newDaemonSet.Name)
			return reconcile.Result{}, err
		}
		// DaemonSet created successfully - return and requeue
		needToRequeue = true
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Rdr DaemonSet")
		return reconcile.Result{}, err
	}

	reqLogger.Info("got Rdr DaemonSet, checking API Ingresses")
	// Check if the Ingresses already exist, if not create new ones
	err = r.reconcileIngress(instance, &needToRequeue)
	if err != nil {
		return reconcile.Result{}, err
	}

	if needToRequeue {
		// one or more resources was created, so requeue the request
		reqLogger.Info("Requeue the request")
		return reconcile.Result{Requeue: true}, nil
	}

	reqLogger.Info("got API Ingresses, checking current DM deployment")
	// Ensure the image is the same as the spec
	var expectedImage string
	if instance.Spec.ImageRegistry == "" {
		expectedImage = res.DefaultImageRegistry + "/" + res.DefaultDmImageName + ":" + res.DefaultDmImageTag
		reqLogger.Info("CS??? default expectedImage=" + expectedImage)
	} else {
		expectedImage = instance.Spec.ImageRegistry + "/" + res.DefaultDmImageName + ":" + res.DefaultDmImageTag
		reqLogger.Info("CS??? expectedImage=" + expectedImage)
	}
	if currentDeployment.Spec.Template.Spec.Containers[0].Image != expectedImage {
		reqLogger.Info("CS??? curr image=" + currentDeployment.Spec.Template.Spec.Containers[0].Image + ", expect=" + expectedImage)
		currentDeployment.Spec.Template.Spec.Containers[0].Image = expectedImage
		reqLogger.Info("updating current DM deployment")
		err = r.client.Update(context.TODO(), currentDeployment)
		if err != nil {
			reqLogger.Error(err, "Failed to update DM Deployment", "Deployment.Namespace", currentDeployment.Namespace,
				"Deployment.Name", currentDeployment.Name)
			return reconcile.Result{}, err
		}
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}

	//CS??? reqLogger.Info("CS??? checking current Rdr DaemonSet")

	reqLogger.Info("Updating Metering status")
	// Update the Metering status with the pod names
	// List the pods for this instance's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Namespace),
		client.MatchingLabels(res.LabelsForSelector(res.DmDeploymentName, meteringCrType, instance.Name)),
	}
	if err = r.client.List(context.TODO(), podList, listOpts...); err != nil {
		reqLogger.Error(err, "Failed to list pods", "Metering.Namespace", instance.Namespace, "Metering.Name", res.DmDeploymentName)
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
			reqLogger.Error(err, "Failed to update Metering status")
			return reconcile.Result{}, err
		}
	}

	reqLogger.Info("CS??? all done")
	return reconcile.Result{}, nil
}

// Check if the DM and Reader Services already exist. If not, create a new one.
// This function was created to reduce the cyclomatic complexity :)
func (r *ReconcileMetering) reconcileService(instance *operatorv1alpha1.Metering, needToRequeue *bool) error {
	reqLogger := log.WithValues("func", "reconcileService", "instance.Name", instance.Name)

	reqLogger.Info("checking DM Service")
	// Check if the DataManager Service already exists, if not create a new one
	dmService := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: res.DmDeploymentName, Namespace: instance.Namespace}, dmService)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Service
		newService := r.serviceForDataMgr(instance)
		reqLogger.Info("Creating a new DM Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
		err = r.client.Create(context.TODO(), newService)
		if err != nil {
			reqLogger.Error(err, "Failed to create new DM Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
			return err
		}
		// Service created successfully - return and requeue
		*needToRequeue = true
	} else if err != nil {
		reqLogger.Error(err, "Failed to get DM Service")
		return err
	}

	reqLogger.Info("got DM Service, checking Rdr Service")
	// Check if the Reader Service already exists, if not create a new one
	readerService := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.ServerServiceName, Namespace: instance.Namespace}, readerService)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Service
		newService := r.serviceForReader(instance)
		reqLogger.Info("Creating a new Rdr Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
		err = r.client.Create(context.TODO(), newService)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Rdr Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
			return err
		}
		// Service created successfully - return and requeue
		*needToRequeue = true
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Rdr Service")
		return err
	}
	reqLogger.Info("got Rdr Service")

	return nil
}

// Check if the Certificates already exist, if not create new ones.
// This function was created to reduce the cyclomatic complexity :)
func (r *ReconcileMetering) reconcileCertificate(instance *operatorv1alpha1.Metering, needToRequeue *bool) error {
	reqLogger := log.WithValues("func", "reconcileCertificate", "instance.Name", instance.Name)

	for _, certData := range certificateList {
		reqLogger.Info("checking Certificate, name=" + certData.Name)
		currentCertificate := &certmgr.Certificate{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: certData.Name, Namespace: instance.Namespace}, currentCertificate)
		if err != nil && errors.IsNotFound(err) {
			// Define a new Certificate
			newCertificate := res.BuildCertificate(instance.Namespace, certData)
			// Set Metering instance as the owner and controller of the Certificate
			err = controllerutil.SetControllerReference(instance, newCertificate, r.scheme)
			if err != nil {
				reqLogger.Error(err, "Failed to set owner for Certificate", "Certificate.Namespace", newCertificate.Namespace,
					"Certificate.Name", newCertificate.Name)
				return err
			}
			reqLogger.Info("Creating a new Certificate", "Certificate.Namespace", newCertificate.Namespace, "Certificate.Name", newCertificate.Name)
			err = r.client.Create(context.TODO(), newCertificate)
			if err != nil {
				reqLogger.Error(err, "Failed to create new Certificate", "Certificate.Namespace", newCertificate.Namespace,
					"Certificate.Name", newCertificate.Name)
				return err
			}
			// Certificate created successfully - return and requeue
			*needToRequeue = true
		} else if err != nil {
			reqLogger.Error(err, "Failed to get Certificate, name="+certData.Name)
			// CertManager might not be installed, so don't fail
			//CS??? return err
		}
	}
	return nil
}

// Check if the Ingresses already exist, if not create new ones.
// This function was created to reduce the cyclomatic complexity :)
func (r *ReconcileMetering) reconcileIngress(instance *operatorv1alpha1.Metering, needToRequeue *bool) error {
	reqLogger := log.WithValues("func", "reconcileIngress", "instance.Name", instance.Name)

	for _, ingressData := range ingressList {
		reqLogger.Info("checking API Ingress, name=" + ingressData.Name)
		currentIngress := &netv1.Ingress{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: ingressData.Name, Namespace: instance.Namespace}, currentIngress)
		if err != nil && errors.IsNotFound(err) {
			// Define a new Ingress
			newIngress := res.BuildIngress(instance.Namespace, ingressData)
			// Set Metering instance as the owner and controller of the Ingress
			err = controllerutil.SetControllerReference(instance, newIngress, r.scheme)
			if err != nil {
				reqLogger.Error(err, "Failed to set owner for API Ingress", "Ingress.Namespace", newIngress.Namespace,
					"Ingress.Name", newIngress.Name)
				return err
			}
			reqLogger.Info("Creating a new API Ingress", "Ingress.Namespace", newIngress.Namespace, "Ingress.Name", newIngress.Name)
			err = r.client.Create(context.TODO(), newIngress)
			if err != nil {
				reqLogger.Error(err, "Failed to create new API Ingress", "Ingress.Namespace", newIngress.Namespace,
					"Ingress.Name", newIngress.Name)
				return err
			}
			// Ingress created successfully - return and requeue
			*needToRequeue = true
		} else if err != nil {
			reqLogger.Error(err, "Failed to get API Ingress, name="+ingressData.Name)
			return err
		}
	}
	return nil
}

// deploymentForDataMgr returns a DataManager Deployment object
func (r *ReconcileMetering) deploymentForDataMgr(instance *operatorv1alpha1.Metering) *appsv1.Deployment {
	reqLogger := log.WithValues("func", "deploymentForDataMgr", "instance.Name", instance.Name)
	metaLabels := res.LabelsForMetadata(res.DmDeploymentName)
	selectorLabels := res.LabelsForSelector(res.DmDeploymentName, meteringCrType, instance.Name)
	podLabels := res.LabelsForPodMetadata(res.DmDeploymentName, meteringCrType, instance.Name)

	var dmImage string
	if instance.Spec.ImageRegistry == "" {
		dmImage = res.DefaultImageRegistry + "/" + res.DefaultDmImageName + ":" + res.DefaultDmImageTag
		reqLogger.Info("CS??? default dmImage=" + dmImage)
	} else {
		dmImage = instance.Spec.ImageRegistry + "/" + res.DefaultDmImageName + ":" + res.DefaultDmImageTag
		reqLogger.Info("CS??? dmImage=" + dmImage)
	}

	dmSecretCheckContainer := res.BaseSecretCheckContainer
	dmSecretCheckContainer.Image = dmImage
	dmSecretCheckContainer.Name = res.DmDeploymentName + "-secret-check"
	// set the SECRET_LIST env var
	dmSecretCheckContainer.Env[res.SecretListVarNdx].Value = res.ReceiverCertZecretName + " " + res.CommonZecretCheckNames
	// set the SECRET_DIR_LIST env var
	dmSecretCheckContainer.Env[res.SecretDirVarNdx].Value = res.ReceiverCertZecretName + " " + res.CommonZecretCheckDirs
	dmSecretCheckContainer.VolumeMounts = append(res.CommonSecretCheckVolumeMounts, res.ReceiverCertVolumeMount)

	dmInitContainer := res.BaseInitContainer
	dmInitContainer.Image = dmImage
	dmInitContainer.Name = res.DmDeploymentName + "-init"
	envVar := corev1.EnvVar{
		Name:  "MCM_VERBOSE",
		Value: "true",
	}
	dmInitContainer.Env = append(dmInitContainer.Env, res.CommonEnvVars...)
	dmInitContainer.Env = append(dmInitContainer.Env, mongoDBEnvVars...)
	dmInitContainer.Env = append(dmInitContainer.Env, envVar)

	dmMainContainer := res.DmMainContainer
	dmMainContainer.Image = dmImage
	dmMainContainer.Name = res.DmDeploymentName
	dmMainContainer.Env = append(dmMainContainer.Env, res.CommonEnvVars...)
	dmMainContainer.Env = append(dmMainContainer.Env, res.IAMEnvVars...)
	dmMainContainer.Env = append(dmMainContainer.Env, mongoDBEnvVars...)
	dmMainContainer.Env = append(dmMainContainer.Env, clusterEnvVars...)
	dmMainContainer.VolumeMounts = append(dmMainContainer.VolumeMounts, res.CommonMainVolumeMounts...)

	dmVolumes := append(commonVolumes, res.ReceiverCertVolume)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.DmDeploymentName,
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
					Volumes: dmVolumes,
					InitContainers: []corev1.Container{
						dmSecretCheckContainer,
						dmInitContainer,
					},
					Containers: []corev1.Container{
						dmMainContainer,
					},
				},
			},
		},
	}
	// Set Metering instance as the owner and controller of the Deployment
	err := controllerutil.SetControllerReference(instance, deployment, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for DM Deployment")
		return nil
	}
	return deployment
}

// serviceForDataMgr returns a DataManager Service object
func (r *ReconcileMetering) serviceForDataMgr(instance *operatorv1alpha1.Metering) *corev1.Service {
	reqLogger := log.WithValues("func", "serviceForDataMgr", "instance.Name", instance.Name)
	metaLabels := res.LabelsForMetadata(res.DmDeploymentName)
	selectorLabels := res.LabelsForSelector(res.DmDeploymentName, meteringCrType, instance.Name)

	reqLogger.Info("CS??? Entry")
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.DmDeploymentName,
			Namespace: instance.Namespace,
			Labels:    metaLabels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "datamanager",
					Port: 3000,
				},
			},
			Selector: selectorLabels,
		},
	}

	// Set Metering instance as the owner and controller of the Service
	err := controllerutil.SetControllerReference(instance, service, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for DM Service")
		return nil
	}
	return service
}

// serviceForReader returns a Reader Service object
func (r *ReconcileMetering) serviceForReader(instance *operatorv1alpha1.Metering) *corev1.Service {
	reqLogger := log.WithValues("func", "serviceForReader", "instance.Name", instance.Name)
	metaLabels := res.LabelsForMetadata(res.ServerServiceName)
	selectorLabels := res.LabelsForSelector(res.ReaderDaemonSetName, meteringCrType, instance.Name)

	reqLogger.Info("CS??? Entry")
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.ServerServiceName,
			Namespace: instance.Namespace,
			Labels:    metaLabels,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name: "apiserver",
					Port: 4000,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 4000,
					},
				},
				{
					Name: "internal-api",
					Port: 4002,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 4002,
					},
				},
			},
			Selector: selectorLabels,
		},
	}

	// Set Metering instance as the owner and controller of the Service
	err := controllerutil.SetControllerReference(instance, service, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for Rdr Service")
		return nil
	}
	return service
}

// daemonForReader returns a Reader DaemonSet object
func (r *ReconcileMetering) daemonForReader(instance *operatorv1alpha1.Metering) *appsv1.DaemonSet {
	reqLogger := log.WithValues("func", "daemonForReader", "instance.Name", instance.Name)
	metaLabels := res.LabelsForMetadata(res.ReaderDaemonSetName)
	selectorLabels := res.LabelsForSelector(res.ReaderDaemonSetName, meteringCrType, instance.Name)
	podLabels := res.LabelsForPodMetadata(res.ReaderDaemonSetName, meteringCrType, instance.Name)

	var image string
	if instance.Spec.ImageRegistry == "" {
		image = res.DefaultImageRegistry + "/" + res.DefaultDmImageName + ":" + res.DefaultDmImageTag
		reqLogger.Info("CS??? default rdrImage=" + image)
	} else {
		image = instance.Spec.ImageRegistry + "/" + res.DefaultDmImageName + ":" + res.DefaultDmImageTag
		reqLogger.Info("CS??? rdrImage=" + image)
	}

	rdrSecretCheckContainer := res.BaseSecretCheckContainer
	rdrSecretCheckContainer.Image = image
	rdrSecretCheckContainer.Name = res.ReaderDaemonSetName + "-secret-check"
	// set the SECRET_LIST env var
	rdrSecretCheckContainer.Env[res.SecretListVarNdx].Value = res.APICertZecretName + " " + res.CommonZecretCheckNames
	// set the SECRET_DIR_LIST env var
	rdrSecretCheckContainer.Env[res.SecretDirVarNdx].Value = res.APICertZecretName + " " + res.CommonZecretCheckDirs
	rdrSecretCheckContainer.VolumeMounts = append(res.CommonSecretCheckVolumeMounts, res.APICertVolumeMount)

	rdrInitContainer := res.BaseInitContainer
	rdrInitContainer.Image = image
	rdrInitContainer.Name = res.ReaderDaemonSetName + "-init"
	rdrInitContainer.Env = append(rdrInitContainer.Env, res.CommonEnvVars...)
	rdrInitContainer.Env = append(rdrInitContainer.Env, mongoDBEnvVars...)

	rdrMainContainer := res.RdrMainContainer
	rdrMainContainer.Image = image
	rdrMainContainer.Name = res.ReaderDaemonSetName
	rdrMainContainer.Env = append(rdrMainContainer.Env, res.CommonEnvVars...)
	rdrMainContainer.Env = append(rdrMainContainer.Env, res.IAMEnvVars...)
	rdrMainContainer.Env = append(rdrMainContainer.Env, mongoDBEnvVars...)
	rdrMainContainer.Env = append(rdrMainContainer.Env, clusterEnvVars...)
	rdrMainContainer.VolumeMounts = append(rdrMainContainer.VolumeMounts, res.CommonMainVolumeMounts...)

	rdrVolumes := append(commonVolumes, res.APICertVolume)

	daemon := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.ReaderDaemonSetName,
			Namespace: instance.Namespace,
			Labels:    metaLabels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: selectorLabels,
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
					Labels: podLabels,
				},
				Spec: corev1.PodSpec{
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
					Volumes: rdrVolumes,
					InitContainers: []corev1.Container{
						rdrSecretCheckContainer,
						rdrInitContainer,
					},
					Containers: []corev1.Container{
						rdrMainContainer,
					},
				},
			},
		},
	}

	// Set Metering instance as the owner and controller of the DaemonSet
	err := controllerutil.SetControllerReference(instance, daemon, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for Rdr DaemonSet")
		return nil
	}
	return daemon
}
