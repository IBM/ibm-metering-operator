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
	"time"

	"k8s.io/apimachinery/pkg/util/intstr"

	operatorv1alpha1 "github.com/ibm/ibm-metering-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-metering-operator/pkg/resources"
	certmgr "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"

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
	/* CS??? disable this code in case CertManager is not installed
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

	version := instance.Spec.Version
	reqLogger.Info("got Metering instance, version=" + version)
	reqLogger.Info("Checking Services")
	// Check if the DM, Reader and Receiver Services already exist. If not, create new ones.
	err = r.reconcileService(instance, &needToRequeue)
	if err != nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("Checking Certificates")
	// Check if the Certificates already exist, if not create new ones
	err = r.reconcileCertificate(instance, &needToRequeue)
	if err != nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("Checking DM Deployment")
	// set common MongoDB env vars based on the instance
	mongoDBEnvVars = res.BuildMongoDBEnvVars(instance.Spec.MongoDB.Host, instance.Spec.MongoDB.Port,
		instance.Spec.MongoDB.UsernameSecret, instance.Spec.MongoDB.UsernameKey,
		instance.Spec.MongoDB.PasswordSecret, instance.Spec.MongoDB.PasswordKey)
	// set common cluster env vars based on the instance
	clusterEnvVars = res.BuildCommonClusterEnvVars(instance.Namespace, instance.Spec.IAMnamespace,
		instance.Spec.External.ClusterName, res.ClusterNameVar)

	// set common Volumes based on the instance
	commonVolumes = res.BuildCommonVolumes(instance.Spec.MongoDB.ClusterCertsSecret, instance.Spec.MongoDB.ClientCertsSecret,
		instance.Spec.MongoDB.UsernameSecret, instance.Spec.MongoDB.PasswordSecret, res.DmDeploymentName, "loglevel")

	// Check if the DM Deployment already exists, if not create a new one
	newDeployment, err := r.deploymentForDataMgr(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	currentDeployment := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.DmDeploymentName, Namespace: instance.Namespace}, currentDeployment)
	if err != nil && errors.IsNotFound(err) {
		// Create a new deployment
		reqLogger.Info("Creating a new DM Deployment", "Deployment.Namespace", newDeployment.Namespace, "Deployment.Name", newDeployment.Name)
		err = r.client.Create(context.TODO(), newDeployment)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue
			reqLogger.Info("DM Deployment already exists")
			needToRequeue = true
		} else if err != nil {
			reqLogger.Error(err, "Failed to create new DM Deployment", "Deployment.Namespace", newDeployment.Namespace,
				"Deployment.Name", newDeployment.Name)
			return reconcile.Result{}, err
		} else {
			// Deployment created successfully - return and requeue
			needToRequeue = true
		}
	} else if err != nil {
		reqLogger.Error(err, "Failed to get DM Deployment")
		return reconcile.Result{}, err
	} else {
		// Found deployment, so send an update to k8s and let it determine if the resource has changed
		reqLogger.Info("Updating DM Deployment")
		currentDeployment.Spec = newDeployment.Spec
		err = r.client.Update(context.TODO(), currentDeployment)
		if err != nil {
			reqLogger.Error(err, "Failed to update DM Deployment", "Deployment.Namespace", currentDeployment.Namespace,
				"Deployment.Name", currentDeployment.Name)
			return reconcile.Result{}, err
		}
	}

	reqLogger.Info("Checking Reader DaemonSet")
	// Check if the DaemonSet already exists, if not create a new one
	newDaemonSet, err := r.daemonForReader(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	currentDaemonSet := &appsv1.DaemonSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.ReaderDaemonSetName, Namespace: instance.Namespace}, currentDaemonSet)
	if err != nil && errors.IsNotFound(err) {
		// Create a new DaemonSet
		reqLogger.Info("Creating a new Reader DaemonSet", "DaemonSet.Namespace", newDaemonSet.Namespace, "DaemonSet.Name", newDaemonSet.Name)
		err = r.client.Create(context.TODO(), newDaemonSet)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue
			reqLogger.Info("Reader DaemonSet already exists")
			needToRequeue = true
		} else if err != nil {
			reqLogger.Error(err, "Failed to create new Reader DaemonSet", "DaemonSet.Namespace", newDaemonSet.Namespace,
				"DaemonSet.Name", newDaemonSet.Name)
			return reconcile.Result{}, err
		} else {
			// DaemonSet created successfully - return and requeue
			needToRequeue = true
		}
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Reader DaemonSet")
		return reconcile.Result{}, err
	} else {
		// Found DaemonSet, so send an update to k8s and let it determine if the resource has changed
		reqLogger.Info("Updating Reader DaemonSet")
		currentDaemonSet.Spec = newDaemonSet.Spec
		err = r.client.Update(context.TODO(), currentDaemonSet)
		if err != nil {
			reqLogger.Error(err, "Failed to update Reader DaemonSet", "DaemonSet.Namespace", currentDaemonSet.Namespace,
				"DaemonSet.Name", currentDaemonSet.Name)
			return reconcile.Result{}, err
		}
	}

	reqLogger.Info("Checking API Ingresses")
	// Check if the Ingresses already exist, if not create new ones
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

	reqLogger.Info("Updating Metering status")
	// Update the Metering status with the pod names.
	// List the pods for this instance's Deployment and DaemonSet
	podNames, err := r.getAllPodNames(instance)
	if err != nil {
		reqLogger.Error(err, "Failed to list pods")
		return reconcile.Result{}, err
	}

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, instance.Status.Nodes) {
		instance.Status.Nodes = podNames
		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update Metering status")
			return reconcile.Result{}, err
		}
	}

	reqLogger.Info("Reconciliation completed")
	// since we updated the status in the Metering CR, sleep 5 seconds to allow the CR to be refreshed.
	time.Sleep(5 * time.Second)
	return reconcile.Result{}, nil
}

// Check if the DM and Reader Services already exist. If not, create new ones.
// This function was created to reduce the cyclomatic complexity :)
func (r *ReconcileMetering) reconcileService(instance *operatorv1alpha1.Metering, needToRequeue *bool) error {
	reqLogger := log.WithValues("func", "reconcileService", "instance.Name", instance.Name)

	reqLogger.Info("Checking DM Service")
	// Check if the DataManager Service already exists, if not create a new one
	newDmService, err := r.serviceForDataMgr(instance)
	if err != nil {
		return err
	}
	err = r.reconcileOneService(instance.Namespace, res.DmServiceName, "DM", newDmService, needToRequeue)
	if err != nil {
		return err
	}

	reqLogger.Info("Checking Reader Service")
	// Check if the Reader Service already exists, if not create a new one
	newRdrService, err := r.serviceForReader(instance)
	if err != nil {
		return err
	}
	err = r.reconcileOneService(instance.Namespace, res.ReaderServiceName, "Reader", newRdrService, needToRequeue)
	if err != nil {
		return err
	}

	if instance.Spec.MultiCloudReceiverEnabled {
		reqLogger.Info("Checking Receiver Service")
		// Check if the Receiver Service already exists, if not create a new one
		newReceiverService, err := r.serviceForReceiver(instance)
		if err != nil {
			return err
		}
		err = r.reconcileOneService(instance.Namespace, res.ReceiverServiceName, "Receiver", newReceiverService, needToRequeue)
		if err != nil {
			return err
		}
	}

	return nil
}

// Check if a Service already exists. If not, create a new one.
// This function was created to reduce the cyclomatic complexity :)
func (r *ReconcileMetering) reconcileOneService(instanceNamespace, serviceName, serviceType string,
	newService *corev1.Service, needToRequeue *bool) error {

	reqLogger := log.WithValues("func", "reconcileOneService", "serviceName", serviceName)

	currentService := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: serviceName, Namespace: instanceNamespace}, currentService)
	if err != nil && errors.IsNotFound(err) {
		// Create a new Service
		reqLogger.Info("Creating a new "+serviceType+" Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
		err = r.client.Create(context.TODO(), newService)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue
			reqLogger.Info(serviceType + " Service already exists")
			*needToRequeue = true
		} else if err != nil {
			reqLogger.Error(err, "Failed to create new "+serviceType+" Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
			return err
		} else {
			// Service created successfully - return and requeue
			*needToRequeue = true
		}
	} else if err != nil {
		reqLogger.Error(err, "Failed to get "+serviceType+" Service")
		return err
	} else {
		// Found service, so send an update to k8s and let it determine if the resource has changed
		reqLogger.Info("Updating " + serviceType + " Service")
		// Can't copy the entire Spec because ClusterIP is immutable
		currentService.Spec.Ports = newService.Spec.Ports
		currentService.Spec.Selector = newService.Spec.Selector
		err = r.client.Update(context.TODO(), currentService)
		if err != nil {
			reqLogger.Error(err, "Failed to update "+serviceType+" Service", "Deployment.Namespace", currentService.Namespace,
				"Deployment.Name", currentService.Name)
			return err
		}
	}

	return nil
}

// Check if the Certificates already exist, if not create new ones.
// This function was created to reduce the cyclomatic complexity :)
func (r *ReconcileMetering) reconcileCertificate(instance *operatorv1alpha1.Metering, needToRequeue *bool) error {
	reqLogger := log.WithValues("func", "reconcileCertificate", "instance.Name", instance.Name)

	if instance.Spec.MultiCloudReceiverEnabled {
		// need to create the receiver certificate
		certificateList = append(certificateList, res.ReceiverCertificateData)
	}
	for _, certData := range certificateList {
		reqLogger.Info("Checking Certificate, name=" + certData.Name)
		newCertificate := res.BuildCertificate(instance.Namespace, certData)
		// Set Metering instance as the owner and controller of the Certificate
		err := controllerutil.SetControllerReference(instance, newCertificate, r.scheme)
		if err != nil {
			reqLogger.Error(err, "Failed to set owner for Certificate", "Certificate.Namespace", newCertificate.Namespace,
				"Certificate.Name", newCertificate.Name)
			return err
		}
		currentCertificate := &certmgr.Certificate{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: certData.Name, Namespace: instance.Namespace}, currentCertificate)
		if err != nil && errors.IsNotFound(err) {
			// Create a new Certificate
			reqLogger.Info("Creating a new Certificate", "Certificate.Namespace", newCertificate.Namespace, "Certificate.Name", newCertificate.Name)
			err = r.client.Create(context.TODO(), newCertificate)
			if err != nil && errors.IsAlreadyExists(err) {
				// Already exists from previous reconcile, requeue
				reqLogger.Info("Certificate already exists")
				*needToRequeue = true
			} else if err != nil {
				reqLogger.Error(err, "Failed to create new Certificate", "Certificate.Namespace", newCertificate.Namespace,
					"Certificate.Name", newCertificate.Name)
				return err
			} else {
				// Certificate created successfully - return and requeue
				*needToRequeue = true
			}
		} else if err != nil {
			reqLogger.Error(err, "Failed to get Certificate, name="+certData.Name)
			// CertManager might not be installed, so don't fail
			//CS??? return err
		} else {
			// Found Certificate, so send an update to k8s and let it determine if the resource has changed
			reqLogger.Info("Updating Certificate", "Certificate.Name", newCertificate.Name)
			currentCertificate.Spec = newCertificate.Spec
			err = r.client.Update(context.TODO(), currentCertificate)
			if err != nil {
				reqLogger.Error(err, "Failed to update Certificate", "Certificate.Namespace", newCertificate.Namespace,
					"Certificate.Name", newCertificate.Name)
				return err
			}
		}
	}
	return nil
}

// Check if the Ingresses already exist, if not create new ones.
// This function was created to reduce the cyclomatic complexity :)
func (r *ReconcileMetering) reconcileIngress(instance *operatorv1alpha1.Metering, needToRequeue *bool) error {
	reqLogger := log.WithValues("func", "reconcileIngress", "instance.Name", instance.Name)

	for _, ingressData := range ingressList {
		reqLogger.Info("Checking API Ingress, name=" + ingressData.Name)
		newIngress := res.BuildIngress(instance.Namespace, ingressData)
		// Set Metering instance as the owner and controller of the Ingress
		err := controllerutil.SetControllerReference(instance, newIngress, r.scheme)
		if err != nil {
			reqLogger.Error(err, "Failed to set owner for API Ingress", "Ingress.Namespace", newIngress.Namespace,
				"Ingress.Name", newIngress.Name)
			return err
		}
		currentIngress := &netv1.Ingress{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: ingressData.Name, Namespace: instance.Namespace}, currentIngress)
		if err != nil && errors.IsNotFound(err) {
			// Create a new Ingress
			reqLogger.Info("Creating a new API Ingress", "Ingress.Namespace", newIngress.Namespace, "Ingress.Name", newIngress.Name)
			err = r.client.Create(context.TODO(), newIngress)
			if err != nil && errors.IsAlreadyExists(err) {
				// Already exists from previous reconcile, requeue
				reqLogger.Info("API Ingress already exists")
				*needToRequeue = true
			} else if err != nil {
				reqLogger.Error(err, "Failed to create new API Ingress", "Ingress.Namespace", newIngress.Namespace,
					"Ingress.Name", newIngress.Name)
				return err
			} else {
				// Ingress created successfully - return and requeue
				*needToRequeue = true
			}
		} else if err != nil {
			reqLogger.Error(err, "Failed to get API Ingress, name="+ingressData.Name)
			return err
		} else {
			// Found Ingress, so send an update to k8s and let it determine if the resource has changed
			reqLogger.Info("Updating API Ingress", "Ingress.Name", newIngress.Name)
			currentIngress.Spec = newIngress.Spec
			err = r.client.Update(context.TODO(), currentIngress)
			if err != nil {
				reqLogger.Error(err, "Failed to update API Ingress", "Ingress.Namespace", newIngress.Namespace,
					"Ingress.Name", newIngress.Name)
				return err
			}
		}
	}
	return nil
}

// deploymentForDataMgr returns a DataManager Deployment object
func (r *ReconcileMetering) deploymentForDataMgr(instance *operatorv1alpha1.Metering) (*appsv1.Deployment, error) {
	reqLogger := log.WithValues("func", "deploymentForDataMgr", "instance.Name", instance.Name)
	metaLabels := res.LabelsForMetadata(res.DmDeploymentName)
	selectorLabels := res.LabelsForSelector(res.DmDeploymentName, meteringCrType, instance.Name)
	podLabels := res.LabelsForPodMetadata(res.DmDeploymentName, meteringCrType, instance.Name)

	var dmImage, imageRegistry string
	if instance.Spec.ImageRegistry == "" {
		imageRegistry = res.DefaultImageRegistry
		reqLogger.Info("use default imageRegistry=" + imageRegistry)
	} else {
		imageRegistry = instance.Spec.ImageRegistry
		reqLogger.Info("use instance imageRegistry=" + imageRegistry)
	}
	dmImage = imageRegistry + "/" + res.DefaultDmImageName + ":" + res.DefaultDmImageTag + instance.Spec.ImageTagPostfix
	reqLogger.Info("dmImage=" + dmImage)

	volumeMounts := res.CommonSecretCheckVolumeMounts
	var nameList, dirList string
	if instance.Spec.MultiCloudReceiverEnabled {
		// set the SECRET_LIST env var
		nameList = res.ReceiverCertSecretName + " " + res.CommonSecretCheckNames
		// set the SECRET_DIR_LIST env var
		dirList = res.ReceiverCertSecretName + " " + res.CommonSecretCheckDirs
		// add the volume mount for the receiver cert
		volumeMounts = append(volumeMounts, res.ReceiverCertVolumeMountForSecretCheck)
	} else {
		nameList = res.CommonSecretCheckNames
		// set the SECRET_DIR_LIST env var
		dirList = res.CommonSecretCheckDirs
	}

	dmSecretCheckContainer := res.BuildSecretCheckContainer(res.DmDeploymentName, dmImage,
		res.SecretCheckCmd, nameList, dirList, volumeMounts)

	initEnvVars := []corev1.EnvVar{
		{
			Name:  "MCM_VERBOSE",
			Value: "true",
		},
	}
	initEnvVars = append(initEnvVars, res.CommonEnvVars...)
	initEnvVars = append(initEnvVars, mongoDBEnvVars...)
	dmInitContainer := res.BuildInitContainer(res.DmDeploymentName, dmImage, initEnvVars)

	receiverEnvVars := res.BuildReceiverEnvVars(instance.Spec.MultiCloudReceiverEnabled)
	dmMainContainer := res.DmMainContainer
	dmMainContainer.Image = dmImage
	dmMainContainer.Name = res.DmDeploymentName
	dmMainContainer.Env = append(dmMainContainer.Env, receiverEnvVars...)
	dmMainContainer.Env = append(dmMainContainer.Env, res.IAMEnvVars...)
	dmMainContainer.Env = append(dmMainContainer.Env, clusterEnvVars...)
	dmMainContainer.Env = append(dmMainContainer.Env, res.CommonEnvVars...)
	dmMainContainer.Env = append(dmMainContainer.Env, mongoDBEnvVars...)

	dmVolumes := commonVolumes
	if instance.Spec.MultiCloudReceiverEnabled {
		dmMainContainer.VolumeMounts = append(dmMainContainer.VolumeMounts, res.ReceiverCertVolumeMountForMain)
		dmVolumes = append(dmVolumes, res.ReceiverCertVolume)
	}
	dmMainContainer.VolumeMounts = append(dmMainContainer.VolumeMounts, res.CommonMainVolumeMounts...)

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
		return nil, err
	}
	return deployment, nil
}

// serviceForDataMgr returns a DataManager Service object
func (r *ReconcileMetering) serviceForDataMgr(instance *operatorv1alpha1.Metering) (*corev1.Service, error) {
	reqLogger := log.WithValues("func", "serviceForDataMgr", "instance.Name", instance.Name)
	metaLabels := res.LabelsForMetadata(res.DmDeploymentName)
	selectorLabels := res.LabelsForSelector(res.DmDeploymentName, meteringCrType, instance.Name)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.DmServiceName,
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
		return nil, err
	}
	return service, nil
}

// serviceForReader returns a Reader Service object
func (r *ReconcileMetering) serviceForReader(instance *operatorv1alpha1.Metering) (*corev1.Service, error) {
	reqLogger := log.WithValues("func", "serviceForReader", "instance.Name", instance.Name)
	metaLabels := res.LabelsForMetadata(res.ReaderServiceName)
	selectorLabels := res.LabelsForSelector(res.ReaderDaemonSetName, meteringCrType, instance.Name)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.ReaderServiceName,
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
		reqLogger.Error(err, "Failed to set owner for Reader Service")
		return nil, err
	}
	return service, nil
}

// serviceForReceiver returns a Receiver Service object
func (r *ReconcileMetering) serviceForReceiver(instance *operatorv1alpha1.Metering) (*corev1.Service, error) {
	reqLogger := log.WithValues("func", "serviceForReceiver", "instance.Name", instance.Name)
	metaLabels := res.LabelsForMetadata(res.DmDeploymentName)
	selectorLabels := res.LabelsForSelector(res.DmDeploymentName, meteringCrType, instance.Name)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.ReceiverServiceName,
			Namespace: instance.Namespace,
			Labels:    metaLabels,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name: "metering-receiver",
					Port: 5000,
				},
			},
			Selector: selectorLabels,
		},
	}

	// Set Metering instance as the owner and controller of the Service
	err := controllerutil.SetControllerReference(instance, service, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for Receiver Service")
		return nil, err
	}
	return service, nil
}

// daemonForReader returns a Reader DaemonSet object
func (r *ReconcileMetering) daemonForReader(instance *operatorv1alpha1.Metering) (*appsv1.DaemonSet, error) {
	reqLogger := log.WithValues("func", "daemonForReader", "instance.Name", instance.Name)
	metaLabels := res.LabelsForMetadata(res.ReaderDaemonSetName)
	selectorLabels := res.LabelsForSelector(res.ReaderDaemonSetName, meteringCrType, instance.Name)
	podLabels := res.LabelsForPodMetadata(res.ReaderDaemonSetName, meteringCrType, instance.Name)

	var rdrImage, imageRegistry string
	if instance.Spec.ImageRegistry == "" {
		imageRegistry = res.DefaultImageRegistry
		reqLogger.Info("use default imageRegistry=" + imageRegistry)
	} else {
		imageRegistry = instance.Spec.ImageRegistry
		reqLogger.Info("use instance imageRegistry=" + imageRegistry)
	}
	rdrImage = imageRegistry + "/" + res.DefaultDmImageName + ":" + res.DefaultDmImageTag + instance.Spec.ImageTagPostfix
	reqLogger.Info("rdrImage=" + rdrImage)

	// set the SECRET_LIST env var
	nameList := res.APICertSecretName + " " + res.CommonSecretCheckNames
	// set the SECRET_DIR_LIST env var
	dirList := res.APICertSecretName + " " + res.CommonSecretCheckDirs
	volumeMounts := append(res.CommonSecretCheckVolumeMounts, res.APICertVolumeMount)
	rdrSecretCheckContainer := res.BuildSecretCheckContainer(res.ReaderDaemonSetName, rdrImage,
		res.SecretCheckCmd, nameList, dirList, volumeMounts)

	initEnvVars := []corev1.EnvVar{}
	initEnvVars = append(initEnvVars, res.CommonEnvVars...)
	initEnvVars = append(initEnvVars, mongoDBEnvVars...)
	rdrInitContainer := res.BuildInitContainer(res.ReaderDaemonSetName, rdrImage, initEnvVars)

	rdrMainContainer := res.RdrMainContainer
	rdrMainContainer.Image = rdrImage
	rdrMainContainer.Name = res.ReaderDaemonSetName
	rdrMainContainer.Env = append(rdrMainContainer.Env, res.IAMEnvVars...)
	rdrMainContainer.Env = append(rdrMainContainer.Env, clusterEnvVars...)
	rdrMainContainer.Env = append(rdrMainContainer.Env, res.CommonEnvVars...)
	rdrMainContainer.Env = append(rdrMainContainer.Env, mongoDBEnvVars...)
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
		reqLogger.Error(err, "Failed to set owner for Reader DaemonSet")
		return nil, err
	}
	return daemon, nil
}

// getAllPodNames returns the pod names of the array of pods passed in
func (r *ReconcileMetering) getAllPodNames(instance *operatorv1alpha1.Metering) ([]string, error) {
	reqLogger := log.WithValues("func", "getAllPodNames")
	// List the pods for this instance's Deployment
	dmPodList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Namespace),
		client.MatchingLabels(res.LabelsForSelector(res.DmDeploymentName, meteringCrType, instance.Name)),
	}
	if err := r.client.List(context.TODO(), dmPodList, listOpts...); err != nil {
		reqLogger.Error(err, "Failed to list pods", "Metering.Namespace", instance.Namespace, "Deployment.Name", res.DmDeploymentName)
		return nil, err
	}
	// List the pods for this instance's DaemonSet
	rdrPodList := &corev1.PodList{}
	listOpts = []client.ListOption{
		client.InNamespace(instance.Namespace),
		client.MatchingLabels(res.LabelsForSelector(res.ReaderDaemonSetName, meteringCrType, instance.Name)),
	}
	if err := r.client.List(context.TODO(), rdrPodList, listOpts...); err != nil {
		reqLogger.Error(err, "Failed to list pods", "Metering.Namespace", instance.Namespace, "DaemonSet.Name", res.ReaderDaemonSetName)
		return nil, err
	}

	podNames := res.GetPodNames(dmPodList.Items)
	podNames = append(podNames, res.GetPodNames(rdrPodList.Items)...)

	return podNames, nil
}
