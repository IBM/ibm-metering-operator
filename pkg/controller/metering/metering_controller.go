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
	"strconv"
	"time"

	operatorv1alpha1 "github.com/ibm/ibm-metering-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-metering-operator/pkg/resources"
	mversion "github.com/ibm/ibm-metering-operator/version"
	certmgr "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"

	ocpoperatorv1 "github.com/openshift/api/operator/v1"
	ocproutev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/extensions/v1beta1"
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

const meteringCrType = "metering_cr"

var commonVolumes = []corev1.Volume{}

var mongoDBEnvVars = []corev1.EnvVar{}
var clusterEnvVars = []corev1.EnvVar{}

var ingressList = []res.IngressData{
	res.APIcheckIngressData,
	res.APIrbacIngressData,
	res.APIswaggerIngressData,
}

var log = logf.Log.WithName("controller_metering")

var isRACMHub = false
var routeHost = ""

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
	reqLogger := log.WithValues("func", "add")

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

	// Watch for changes to secondary resource "Ingress" and requeue the owner Metering
	err = c.Watch(&source.Kind{Type: &netv1.Ingress{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.Metering{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource "Certificate" and requeue the owner Metering
	err = c.Watch(&source.Kind{Type: &certmgr.Certificate{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.Metering{},
	})
	if err != nil {
		reqLogger.Error(err, "Failed to watch Certificate")
		// CertManager might not be installed, so don't fail
		//CS??? return err
	}

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
	reqLogger := log.WithValues("Request.Name", request.Name)
	reqLogger.Info("Reconciling Metering", "Request.Namespace", request.Namespace)

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

	// set a default Status value
	if len(instance.Status.PodNames) == 0 {
		instance.Status.PodNames = res.DefaultStatusForCR
		instance.Status.Versions.Reconciled = mversion.Version
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to set Metering default status")
			return reconcile.Result{}, err
		}
	}

	//If the metering receiver is enabled then find the app domain used for route and certificate generation
	//and determine if this is a RACM hub
	// get appDomain
	isRACMHub = false
	routeHost = ""
	if instance.Spec.MultiCloudReceiverEnabled {
		ing := &ocpoperatorv1.IngressController{}
		namespace := types.NamespacedName{Name: "default", Namespace: "openshift-ingress-operator"}
		err := r.client.Get(context.TODO(), namespace, ing)
		if err != nil {
			reqLogger.Error(err, "Failed to get IngressController")
		} else {
			appDomain := ing.Status.Domain
			if len(appDomain) > 0 {
				routeHost = appDomain
			}
		}

		//Check RHACM
		rhacmErr := res.CheckRhacm(r.client)
		if rhacmErr == nil {
			// multiclusterhub found, this means RHACM exists
			isRACMHub = true
		}
	}

	reqLogger.Info("Checking Services")
	// Check if the DM, Reader and Receiver Services already exist. If not, create new ones.
	err = r.reconcileAllServices(instance, &needToRequeue)
	if err != nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("Checking DM Deployment", "Deployment.Name", res.DmDeploymentName)
	// set common MongoDB env vars based on the instance
	mongoDBEnvVars = res.BuildMongoDBEnvVars(instance.Spec.MongoDB)
	// set common cluster env vars based on the instance
	clusterEnvVars = res.BuildCommonClusterEnvVars(instance.Namespace, instance.Spec.IAMnamespace)

	// set common Volumes based on the instance
	commonVolumes = res.BuildCommonVolumes(instance.Spec.MongoDB, res.DmDeploymentName, "loglevel")

	// Check if the DM Deployment already exists, if not create a new one
	newDmDeployment, err := r.deploymentForDataMgr(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = res.ReconcileDeployment(r.client, instance.Namespace, res.DmDeploymentName, "DM", newDmDeployment, &needToRequeue)
	if err != nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("Checking Reader Deployment", "Deployment.Name", res.ReaderDeploymentName)
	// Check if the Reader Deployment already exists, if not create a new one
	newRdrDeployment, err := r.deploymentForReader(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = res.ReconcileDeployment(r.client, instance.Namespace, res.ReaderDeploymentName, "Reader", newRdrDeployment, &needToRequeue)
	if err != nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("Checking API Ingresses")
	// Check if the Ingresses already exist, if not create new ones
	err = r.reconcileAllIngress(instance, &needToRequeue)
	if err != nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("Checking Certificates")
	// Check if the Certificates already exist, if not create new ones
	err = r.reconcileAllCertificates(instance, &needToRequeue)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Prior to version 3.6, metering-reader ran as a daemon set instead of a deployment.
	// Delete the DaemonSet if it was leftover from a previous version.
	r.checkDaemonSet(instance)

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
	// List the pods for this instance's Deployments
	podNames, err := r.getAllPodNames(instance)
	if err != nil {
		reqLogger.Error(err, "Failed to list pods")
		return reconcile.Result{}, err
	}
	// if no pods were found set the default status
	if len(podNames) == 0 {
		podNames = res.DefaultStatusForCR
	}

	// Update status if needed
	if !reflect.DeepEqual(podNames, instance.Status.PodNames) || (mversion.Version != instance.Status.Versions.Reconciled) {
		instance.Status.PodNames = podNames
		instance.Status.Versions.Reconciled = mversion.Version
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
func (r *ReconcileMetering) reconcileAllServices(instance *operatorv1alpha1.Metering, needToRequeue *bool) error {
	reqLogger := log.WithValues("func", "reconcileAllServices")

	reqLogger.Info("Checking DM Service", "Service.Name", res.DmServiceName)
	// Check if the DataManager Service already exists, if not create a new one
	newDmService, err := r.serviceForDataMgr(instance)
	if err != nil {
		return err
	}
	err = res.ReconcileService(r.client, instance.Namespace, res.DmServiceName, "DM", newDmService, needToRequeue)
	if err != nil {
		return err
	}

	reqLogger.Info("Checking Reader Service", "Service.Name", res.ReaderServiceName)
	// Check if the Reader Service already exists, if not create a new one
	newRdrService, err := r.serviceForReader(instance)
	if err != nil {
		return err
	}
	err = res.ReconcileService(r.client, instance.Namespace, res.ReaderServiceName, "Reader", newRdrService, needToRequeue)
	if err != nil {
		return err
	}

	if instance.Spec.MultiCloudReceiverEnabled {
		reqLogger.Info("Checking Receiver Service", "Service.Name", res.ReceiverServiceName)
		// Check if the Receiver Service already exists, if not create a new one
		newReceiverService, err := r.serviceForReceiver(instance)
		if err != nil {
			return err
		}
		err = res.ReconcileService(r.client, instance.Namespace, res.ReceiverServiceName, "Receiver", newReceiverService, needToRequeue)
		if err != nil {
			return err
		}
		if isRACMHub {
			reqLogger.Info("Checking Receiver Route", "Route.Name", res.ReceiverRouteName)
			// Check if the Receiver Route already exists, if not create a new one
			newReceiverRoute, err := r.routeForReceiver(instance)
			if err != nil {
				return err
			}
			err = res.ReconcileRoute(r.client, instance.Namespace, res.ReceiverRouteName, "Route", newReceiverRoute, needToRequeue)
			if err != nil {
				return err
			}
		}
	} else {
		//If the multicloud receiver is disabled then make sure the service and route to it are deleted
		receiverService := &corev1.Service{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: res.ReceiverServiceName, Namespace: instance.Namespace}, receiverService)
		if err == nil {
			//found the service, it should be removed
			err := r.client.Delete(context.TODO(), receiverService)
			if err != nil {
				reqLogger.Error(err, "Failed to delete receiver Service after the receiver was disabled")
			} else {
				reqLogger.Info("Deleted receiver Service after the receiver was disabled")
			}
		} else if !errors.IsNotFound(err) {
			//Report error if not found
			reqLogger.Error(err, "Failed to delete the receiver Service")
		}

		//If the multicloud receiver is disabled then make sure the route has also been deleted
		receiverRoute := &ocproutev1.Route{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.ReceiverRouteName, Namespace: instance.Namespace}, receiverRoute)
		if err == nil {
			//found route that should be deleted since the multicloud receiver is disabled
			err := r.client.Delete(context.TODO(), receiverRoute)
			if err != nil {
				reqLogger.Error(err, "Failed to delete receiver Route after the receiver was disabled")
			} else {
				reqLogger.Info("Deleted receiver Route after the receiver was disabled")
			}
		} else if !errors.IsNotFound(err) {
			//Report error if not found
			reqLogger.Error(err, "Failed to delete the receiver Route")
		}
	}

	return nil
}

// Check if the Certificates already exist, if not create new ones.
// This function was created to reduce the cyclomatic complexity :)
func (r *ReconcileMetering) reconcileAllCertificates(instance *operatorv1alpha1.Metering, needToRequeue *bool) error {
	reqLogger := log.WithValues("func", "reconcileAllCertificates")

	certificateList := []res.CertificateData{
		res.APICertificateData,
	}
	if instance.Spec.MultiCloudReceiverEnabled {
		// need to create the receiver certificate
		certificateList = append(certificateList, res.ReceiverCertificateData)
	}
	for _, certData := range certificateList {
		reqLogger.Info("Checking Certificate", "Certificate.Name", certData.Name)
		newCertificate := res.BuildCertificate(instance.Namespace, instance.Spec.ClusterIssuer, certData)

		// add a new dnsname for certificate, this will be used for sender
		if newCertificate.Name == res.ReceiverCertName {
			if len(routeHost) > 0 {
				newCertificate.Spec.DNSNames = append(newCertificate.Spec.DNSNames, "metering-receiver."+routeHost)
			} else {
				reqLogger.Info("WARNING: Hub route host for metering receiver certificate is not set - data delivery to the hub will not be possible")
			}
		}

		// Set Metering instance as the owner and controller of the Certificate
		err := controllerutil.SetControllerReference(instance, newCertificate, r.scheme)
		if err != nil {
			reqLogger.Error(err, "Failed to set owner for Certificate", "Certificate.Namespace", newCertificate.Namespace,
				"Certificate.Name", newCertificate.Name)
			return err
		}
		err = res.ReconcileCertificate(r.client, instance.Namespace, certData.Name, newCertificate, needToRequeue)
		if err != nil {
			return err
		}
	}

	//If the multicloud receiver is disabled, ensure that receiver certificate is gone
	if !instance.Spec.MultiCloudReceiverEnabled {
		receiverCert := &certmgr.Certificate{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: res.ReceiverCertificateData.Name, Namespace: instance.Namespace}, receiverCert)
		if err == nil {
			//found the cert, it should be removed
			err := r.client.Delete(context.TODO(), receiverCert)
			if err != nil {
				reqLogger.Error(err, "Failed to delete receiver Certificate after the receiver was disabled")
			} else {
				reqLogger.Info("Deleted receiver Certificate after the receiver was disabled")
			}
		} else if !errors.IsNotFound(err) {
			//Report error if not found
			reqLogger.Error(err, "Failed to delete the receiver Certificate")
		}

		//also delete the secret
		receiverCertSecret := &corev1.Secret{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.ReceiverCertificateData.Secret, Namespace: instance.Namespace}, receiverCertSecret)
		if err == nil {
			//found the cert secret, it should be removed
			err := r.client.Delete(context.TODO(), receiverCertSecret)
			if err != nil {
				reqLogger.Error(err, "Failed to delete receiver Certificate Secret after the receiver was disabled")
			} else {
				reqLogger.Info("Deleted receiver Certificate Secret after the receiver was disabled")
			}
		} else if !errors.IsNotFound(err) {
			//Report error if not found
			reqLogger.Error(err, "Failed to delete the receiver Certificate Secret")
		}

	}

	return nil
}

// Check if the Ingresses already exist, if not create new ones.
// This function was created to reduce the cyclomatic complexity :)
func (r *ReconcileMetering) reconcileAllIngress(instance *operatorv1alpha1.Metering, needToRequeue *bool) error {
	reqLogger := log.WithValues("func", "reconcileAllIngress")

	for _, ingressData := range ingressList {
		reqLogger.Info("Checking API Ingress", "Ingress.Name", ingressData.Name)
		newIngress := res.BuildIngress(instance.Namespace, ingressData)
		// Set Metering instance as the owner and controller of the Ingress
		err := controllerutil.SetControllerReference(instance, newIngress, r.scheme)
		if err != nil {
			reqLogger.Error(err, "Failed to set owner for API Ingress", "Ingress.Namespace", newIngress.Namespace,
				"Ingress.Name", newIngress.Name)
			return err
		}
		err = res.ReconcileIngress(r.client, instance.Namespace, ingressData.Name, "API", newIngress, needToRequeue)
		if err != nil {
			return err
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

	dmImage := res.GetImageID(instance.Spec.ImageRegistry, instance.Spec.ImageTagPostfix,
		res.DefaultImageRegistry, res.DefaultDmImageName, res.VarImageSHAforDM, res.DefaultDmImageTag)
	reqLogger.Info("dmImage=" + dmImage)

	var additionalInfo res.SecretCheckData
	var additionalInfoPtr *res.SecretCheckData
	if instance.Spec.MultiCloudReceiverEnabled {
		// add to the SECRET_LIST env var
		additionalInfo.Names = res.ReceiverCertSecretName
		// add to the SECRET_DIR_LIST env var
		additionalInfo.Dirs = res.ReceiverCertDirName
		// add the volume mount for the receiver cert
		additionalInfo.VolumeMounts = []corev1.VolumeMount{res.ReceiverCertVolumeMountForSecretCheck}
		additionalInfoPtr = &additionalInfo
	} else {
		additionalInfoPtr = nil
	}

	// setup the init containers
	dmSecretCheckContainer := res.BuildSecretCheckContainer(res.DmDeploymentName, dmImage,
		res.SecretCheckCmd, instance.Spec.MongoDB, additionalInfoPtr)

	initEnvVars := []corev1.EnvVar{
		{
			Name:  "MCM_VERBOSE",
			Value: "true",
		},
	}
	initEnvVars = append(initEnvVars, res.CommonEnvVars...)
	initEnvVars = append(initEnvVars, mongoDBEnvVars...)
	dmInitContainer := res.BuildInitContainer(res.DmDeploymentName, dmImage, initEnvVars)

	// setup the main container
	receiverEnvVars := res.BuildReceiverEnvVars(instance.Spec.MultiCloudReceiverEnabled)
	dmMainContainer := res.DmMainContainer
	dmMainContainer.Image = dmImage
	dmMainContainer.Name = res.DmDeploymentName
	// setup environment vars
	dmMainContainer.Env = append(dmMainContainer.Env, receiverEnvVars...)
	dmMainContainer.Env = append(dmMainContainer.Env, res.IAMEnvVars...)
	dmMainContainer.Env = append(dmMainContainer.Env, clusterEnvVars...)
	dmMainContainer.Env = append(dmMainContainer.Env, res.CommonEnvVars...)
	dmMainContainer.Env = append(dmMainContainer.Env, mongoDBEnvVars...)

	// setup volumes and volume mounts
	dmVolumes := commonVolumes
	if instance.Spec.MultiCloudReceiverEnabled {
		dmMainContainer.VolumeMounts = append(dmMainContainer.VolumeMounts, res.ReceiverCertVolumeMountForMain)
		dmVolumes = append(dmVolumes, res.ReceiverCertVolume)
	}
	dmMainContainer.VolumeMounts = append(dmMainContainer.VolumeMounts, res.CommonMainVolumeMounts...)

	// setup the resource requirements
	dmMainContainer.Resources = res.BuildResourceRequirements(instance.Spec.DataManager.DataManagerResources.Resources,
		res.DmResourceRequirements)

	// setup the deployment
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
												Values:   res.ArchitectureList,
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
					Name:     "datamanager",
					Protocol: corev1.ProtocolTCP,
					Port:     3000,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 3000,
					},
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
	selectorLabels := res.LabelsForSelector(res.ReaderDeploymentName, meteringCrType, instance.Name)

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
					Name:     "apiserver",
					Protocol: corev1.ProtocolTCP,
					Port:     4000,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 4000,
					},
				},
				{
					Name:     "internal-api",
					Protocol: corev1.ProtocolTCP,
					Port:     4002,
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
					Name:     "metering-receiver",
					Protocol: corev1.ProtocolTCP,
					Port:     5000,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 5000,
					},
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

// routeForReceiver returns a Receiver Route object
func (r *ReconcileMetering) routeForReceiver(instance *operatorv1alpha1.Metering) (*ocproutev1.Route, error) {
	reqLogger := log.WithValues("func", "routeForReceiver", "instance.Name", instance.Name)

	if len(routeHost) == 0 {
		err := errors.NewBadRequest("Hub routeHost is not set so receiver route cannot be created")
		return nil, err
	}

	metaLabels := res.LabelsForMetadata(res.ReceiverRouteName)
	weight := int32(100)

	route := &ocproutev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.ReceiverRouteName,
			Namespace: instance.Namespace,
			Labels:    metaLabels,
		},
		Spec: ocproutev1.RouteSpec{
			Host: "metering-receiver." + routeHost,
			Port: &ocproutev1.RoutePort{
				TargetPort: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "metering-receiver",
				},
			},
			TLS: &ocproutev1.TLSConfig{
				InsecureEdgeTerminationPolicy: "Redirect",
				Termination:                   "passthrough",
			},
			To: ocproutev1.RouteTargetReference{
				Kind:   "Service",
				Name:   "metering-receiver",
				Weight: &weight,
			},
			WildcardPolicy: "None",
		},
	}

	// Set Metering instance as the owner and controller of the Route
	err := controllerutil.SetControllerReference(instance, route, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for Receiver Route")
		return nil, err
	}
	return route, nil
}

// deploymentForReader returns a Reader Deployment object
func (r *ReconcileMetering) deploymentForReader(instance *operatorv1alpha1.Metering) (*appsv1.Deployment, error) {
	reqLogger := log.WithValues("func", "deploymentForReader", "instance.Name", instance.Name)
	metaLabels := res.LabelsForMetadata(res.ReaderDeploymentName)
	selectorLabels := res.LabelsForSelector(res.ReaderDeploymentName, meteringCrType, instance.Name)
	podLabels := res.LabelsForPodMetadata(res.ReaderDeploymentName, meteringCrType, instance.Name)

	// the Reader code is part of the metering-data-manager image
	rdrImage := res.GetImageID(instance.Spec.ImageRegistry, instance.Spec.ImageTagPostfix,
		res.DefaultImageRegistry, res.DefaultDmImageName, res.VarImageSHAforDM, res.DefaultDmImageTag)

	reqLogger.Info("rdrImage=" + rdrImage)

	var additionalInfo res.SecretCheckData
	// add to the SECRET_LIST env var
	additionalInfo.Names = res.APICertSecretName
	// add to the SECRET_DIR_LIST env var
	additionalInfo.Dirs = res.APICertDirName
	// add the volume mount for the API cert
	additionalInfo.VolumeMounts = []corev1.VolumeMount{res.APICertVolumeMount}

	// setup the init containers
	rdrSecretCheckContainer := res.BuildSecretCheckContainer(res.ReaderDeploymentName, rdrImage,
		res.SecretCheckCmd, instance.Spec.MongoDB, &additionalInfo)

	initEnvVars := []corev1.EnvVar{}
	initEnvVars = append(initEnvVars, res.CommonEnvVars...)
	initEnvVars = append(initEnvVars, mongoDBEnvVars...)
	rdrInitContainer := res.BuildInitContainer(res.ReaderDeploymentName, rdrImage, initEnvVars)

	// setup the main container
	rdrMainContainer := res.RdrMainContainer
	rdrMainContainer.Image = rdrImage
	rdrMainContainer.Name = res.ReaderDeploymentName
	// setup environment vars
	excludeNamespacesEnvVar := corev1.EnvVar{
		Name:  "HC_EXCLUDE_SYSTEM_NAMESPACES",
		Value: strconv.FormatBool(instance.Spec.ExcludeSystemNamespaces),
	}
	rdrMainContainer.Env = append(rdrMainContainer.Env, excludeNamespacesEnvVar)
	rdrMainContainer.Env = append(rdrMainContainer.Env, res.IAMEnvVars...)
	rdrMainContainer.Env = append(rdrMainContainer.Env, clusterEnvVars...)
	rdrMainContainer.Env = append(rdrMainContainer.Env, res.CommonEnvVars...)
	rdrMainContainer.Env = append(rdrMainContainer.Env, mongoDBEnvVars...)

	// setup volumes and volume mounts
	rdrMainContainer.VolumeMounts = append(rdrMainContainer.VolumeMounts, res.CommonMainVolumeMounts...)
	rdrVolumes := append(commonVolumes, res.APICertVolume)

	// setup the resource requirements
	rdrMainContainer.Resources = res.BuildResourceRequirements(instance.Spec.Reader.ReaderResources.Resources,
		res.RdrResourceRequirements)

	// setup the deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.ReaderDeploymentName,
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
												Values:   res.ArchitectureList,
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

	// Set Metering instance as the owner and controller of the Deployment
	err := controllerutil.SetControllerReference(instance, deployment, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for Reader Deployment")
		return nil, err
	}
	return deployment, nil
}

// delete the Reader DaemonSet if it was leftover from a previous version
func (r *ReconcileMetering) checkDaemonSet(instance *operatorv1alpha1.Metering) {
	reqLogger := log.WithValues("func", "checkDaemonSet", "instance.Name", instance.Name)
	daemonSet := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.ReaderDaemonSetName,
			Namespace: res.WatchNamespaceV350,
		},
	}
	// check if the DaemonSet exists
	err := r.client.Get(context.TODO(),
		types.NamespacedName{Name: res.ReaderDaemonSetName, Namespace: res.WatchNamespaceV350}, daemonSet)
	if err == nil {
		// found DaemonSet so delete it
		err := r.client.Delete(context.TODO(), daemonSet)
		if err != nil {
			reqLogger.Error(err, "Failed to delete old Reader DaemonSet")
		} else {
			reqLogger.Info("Deleted old Reader DaemonSet")
		}
	} else if !errors.IsNotFound(err) {
		// if err is NotFound do nothing, else print an error msg
		reqLogger.Error(err, "Failed to get old Reader DaemonSet")
	}
}

// getAllPodNames returns the list of pod names for the associated deployments
func (r *ReconcileMetering) getAllPodNames(instance *operatorv1alpha1.Metering) ([]string, error) {
	reqLogger := log.WithValues("func", "getAllPodNames")
	// List the pods for this instance's DM Deployment
	dmPodList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Namespace),
		client.MatchingLabels(res.LabelsForSelector(res.DmDeploymentName, meteringCrType, instance.Name)),
	}
	if err := r.client.List(context.TODO(), dmPodList, listOpts...); err != nil {
		reqLogger.Error(err, "Failed to list pods", "Metering.Namespace", instance.Namespace, "Deployment.Name", res.DmDeploymentName)
		return nil, err
	}
	// List the pods for this instance's Rdr Deployment
	rdrPodList := &corev1.PodList{}
	listOpts = []client.ListOption{
		client.InNamespace(instance.Namespace),
		client.MatchingLabels(res.LabelsForSelector(res.ReaderDeploymentName, meteringCrType, instance.Name)),
	}
	if err := r.client.List(context.TODO(), rdrPodList, listOpts...); err != nil {
		reqLogger.Error(err, "Failed to list pods", "Metering.Namespace", instance.Namespace, "Deployment.Name", res.ReaderDeploymentName)
		return nil, err
	}

	podNames := res.GetPodNames(dmPodList.Items)
	podNames = append(podNames, res.GetPodNames(rdrPodList.Items)...)

	return podNames, nil
}
