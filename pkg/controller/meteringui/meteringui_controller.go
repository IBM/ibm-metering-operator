//
// Copyright 2021 IBM Corporation
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
	"time"

	operatorv1alpha1 "github.com/ibm/ibm-metering-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-metering-operator/pkg/resources"
	mversion "github.com/ibm/ibm-metering-operator/version"
	certmgr "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	reqLogger := log.WithValues("func", "add")

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

	// Watch for changes to secondary resource "Certificate" and requeue the owner Metering
	err = c.Watch(&source.Kind{Type: &certmgr.Certificate{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.MeteringUI{},
	})
	if err != nil {
		reqLogger.Error(err, "Failed to watch Certificate")
		// CertManager might not be installed, so don't fail
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
	reqLogger := log.WithValues("Request.Name", request.Name)
	reqLogger.Info("Reconciling MeteringUI", "Request.Namespace", request.Namespace)

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

	// set a default Status value
	if len(instance.Status.PodNames) == 0 {
		instance.Status.PodNames = res.DefaultStatusForCR
		instance.Status.Versions.Reconciled = mversion.Version
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to set MeteringUI default status")
			return reconcile.Result{}, err
		}
	}

	reqLogger.Info("Checking UI Service", "Service.Name", res.UIServiceName)
	// Check if the UI Service already exists, if not create a new one
	newService, err := r.serviceForUI(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = res.ReconcileService(r.client, instance.Namespace, res.UIServiceName, "UI", newService, &needToRequeue)
	if err != nil {
		return reconcile.Result{}, err
	}

	// set common MongoDB env vars based on the instance
	mongoDBEnvVars = res.BuildMongoDBEnvVars(instance.Spec.MongoDB)
	// set common cluster env vars based on the instance
	clusterEnvVars = res.BuildUIClusterEnvVars(instance.Namespace, instance.Spec.ClusterName, instance.Spec.UI, false)

	// set common Volumes based on the instance
	commonVolumes = res.BuildCommonVolumes(instance.Spec.MongoDB, res.UIDeploymentName, "loglevel")

	reqLogger.Info("Checking UI Deployment", "Deployment.Name", res.UIDeploymentName)
	// Check if the UI Deployment already exists, if not create a new one
	newDeployment, err := r.deploymentForUI(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = res.ReconcileDeployment(r.client, instance.Namespace, res.UIDeploymentName, "UI", newDeployment, &needToRequeue)
	if err != nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("Checking UI Ingress", "Ingress.Name", res.UIIngressData.Name)
	// Check if the Ingress already exists, if not create a new one
	newIngress := res.BuildIngress(instance.Namespace, res.UIIngressData)
	// Set MeteringUI instance as the owner and controller of the Ingress
	err = controllerutil.SetControllerReference(instance, newIngress, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for UI Ingress", "Ingress.Namespace", newIngress.Namespace,
			"Ingress.Name", newIngress.Name)
		return reconcile.Result{}, err
	}
	err = res.ReconcileIngress(r.client, instance.Namespace, res.UIIngressData.Name, "UI", newIngress, &needToRequeue)
	if err != nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("Checking UI Certificates")
	// Check if the Certificates already exist, if not create new ones
	err = r.reconcileAllCertificates(instance, &needToRequeue)
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
	// if no pods were found set the default status
	if len(podNames) == 0 {
		podNames = res.DefaultStatusForCR
	}

	// Update status.PodNames if needed
	if !reflect.DeepEqual(podNames, instance.Status.PodNames) || (mversion.Version != instance.Status.Versions.Reconciled) {
		instance.Status.PodNames = podNames
		instance.Status.Versions.Reconciled = mversion.Version
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

	// the InitContainer code is part of the metering-data-manager image
	initImage := res.GetImageID(instance.Spec.ImageRegistry, instance.Spec.ImageTagPostfix,
		res.DefaultImageRegistry, res.DefaultDmImageName, res.VarImageForDM, res.DefaultDmImageTag)
	reqLogger.Info("initImage=" + initImage)
	uiImage := res.GetImageID(instance.Spec.ImageRegistry, instance.Spec.ImageTagPostfix,
		res.DefaultImageRegistry, res.DefaultUIImageName, res.VarImageForUI, res.DefaultUIImageTag)
	reqLogger.Info("uiImage=" + uiImage)

	// The apikey and OIDC secret names can be set in the CR, but will be ignored
	// We are hardcoding here since the operand is using hardcoded names
	// save me apiKeySecretName := res.DefaultAPIKeySecretName
	// save me platformOidcSecretName := res.DefaultPlatformOidcSecretName

	// Apikey and OIDC secrets are now read dynamically in the UI container and are no longer
	// part of the secret checker
	var additionalInfo res.SecretCheckData
	// add to the SECRET_LIST env var
	additionalInfo.Names = res.UICertSecretName
	// add to the SECRET_DIR_LIST env var
	additionalInfo.Dirs = res.UICertDirName
	// add the cert secret which is only used by the UI today
	additionalInfo.VolumeMounts = append(additionalInfo.VolumeMounts, res.UICertVolumeMountForSecretCheck)

	// setup the init containers
	uiSecretCheckContainer := res.BuildSecretCheckContainer(res.UIDeploymentName, initImage,
		res.SecretCheckCmd, instance.Spec.MongoDB, &additionalInfo)

	initEnvVars := []corev1.EnvVar{}
	initEnvVars = append(initEnvVars, res.CommonEnvVars...)
	initEnvVars = append(initEnvVars, mongoDBEnvVars...)
	uiInitContainer := res.BuildInitContainer(res.UIDeploymentName, initImage, initEnvVars)

	// setup the main container
	uiMainContainer := res.UIMainContainer
	uiMainContainer.Image = uiImage
	uiMainContainer.Name = res.UIDeploymentName
	// setup environment vars
	uiMainContainer.Env = append(uiMainContainer.Env, res.IAMEnvVars...)
	uiMainContainer.Env = append(uiMainContainer.Env, res.UIEnvVars...)
	uiMainContainer.Env = append(uiMainContainer.Env, clusterEnvVars...)
	uiMainContainer.Env = append(uiMainContainer.Env, res.CommonEnvVars...)
	uiMainContainer.Env = append(uiMainContainer.Env, mongoDBEnvVars...)

	// setup volumes and volume mounts
	uiMainContainer.VolumeMounts = append(uiMainContainer.VolumeMounts, res.CommonMainVolumeMounts...)
	uiMainContainer.VolumeMounts = append(uiMainContainer.VolumeMounts, res.UICertVolumeMountForMain)
	uiVolumes := append(commonVolumes, res.UICertVolume)

	// setup the resource requirements
	uiMainContainer.Resources = res.BuildResourceRequirements(instance.Spec.UI.Resources,
		res.UIResourceRequirements)

	// "replicas" should be set in the CR. if it isn't found, use the default value.
	replicas := res.Replica1
	if instance.Spec.Replicas > 0 {
		replicas = instance.Spec.Replicas
	}

	// setup the deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.UIDeploymentName,
			Namespace: instance.Namespace,
			Labels:    metaLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
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
					Affinity:                      res.GetAffinity(true, res.UIDeploymentName),
					Tolerations:                   res.GetTolerations(),
					TopologySpreadConstraints:     res.GetTopologySpreadConstraints(res.UIDeploymentName),
					Volumes:                       uiVolumes,
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
					Name:     "dashboard",
					Protocol: corev1.ProtocolTCP,
					Port:     3130,
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

// Check if the Certificate already exists, if not create a new one.
// This function was created to reduce the cyclomatic complexity :)
func (r *ReconcileMeteringUI) reconcileAllCertificates(instance *operatorv1alpha1.MeteringUI, needToRequeue *bool) error {
	reqLogger := log.WithValues("func", "reconcileAllCertificates")

	certificateList := []res.CertificateData{
		res.UICertificateData,
	}

	for _, certData := range certificateList {
		reqLogger.Info("Checking Certificate", "Certificate.Name", certData.Name)
		newCertificate := res.BuildCertificate(instance.Namespace, "", certData)
		// Set MeteringUI instance as the owner and controller of the Certificate
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
	return nil
}
