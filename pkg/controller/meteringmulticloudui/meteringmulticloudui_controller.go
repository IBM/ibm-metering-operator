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

package meteringmulticloudui

import (
	"context"
	"reflect"
	"time"

	operatorv1alpha1 "github.com/ibm/ibm-metering-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-metering-operator/pkg/resources"
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

const meteringMcmUICrType = "meteringmulticloudui_cr"

var commonVolumes = []corev1.Volume{}

var mongoDBEnvVars = []corev1.EnvVar{}
var clusterEnvVars = []corev1.EnvVar{}

var log = logf.Log.WithName("controller_meteringmcmui")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new MeteringMultiCloudUI Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMeteringMultiCloudUI{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	reqLogger := log.WithValues("func", "add")

	// Create a new controller
	c, err := controller.New("meteringmulticloudui-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource MeteringMultiCloudUI
	err = c.Watch(&source.Kind{Type: &operatorv1alpha1.MeteringMultiCloudUI{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource "Deployment" and requeue the owner Metering
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.MeteringMultiCloudUI{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource "Service" and requeue the owner Metering
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.MeteringMultiCloudUI{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource "Ingress" and requeue the owner Metering
	err = c.Watch(&source.Kind{Type: &netv1.Ingress{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.MeteringMultiCloudUI{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource "Certificate" and requeue the owner Metering
	err = c.Watch(&source.Kind{Type: &certmgr.Certificate{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.MeteringMultiCloudUI{},
	})
	if err != nil {
		reqLogger.Error(err, "Failed to watch Certificate")
		// CertManager might not be installed, so don't fail
	}

	return nil
}

// blank assignment to verify that ReconcileMeteringMultiCloudUI implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileMeteringMultiCloudUI{}

// ReconcileMeteringMultiCloudUI reconciles a MeteringMultiCloudUI object
type ReconcileMeteringMultiCloudUI struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a MeteringMultiCloudUI object and makes changes based on the state read
// and what is in the MeteringMultiCloudUI.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// an MCM UI Deployment and Service for each MeteringMultiCloudUI CR
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMeteringMultiCloudUI) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Name", request.Name)
	reqLogger.Info("Reconciling MeteringMultiCloudUI", "Request.Namespace", request.Namespace)

	// if we need to create several resources, set a flag so we just requeue one time instead of after each create.
	needToRequeue := false

	// Fetch the MeteringMultiCloudUI CR instance
	instance := &operatorv1alpha1.MeteringMultiCloudUI{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("MeteringMultiCloudUI resource not found. Ignoring since object must be deleted")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get MeteringMultiCloudUI CR")
		return reconcile.Result{}, err
	}

	version := instance.Spec.Version
	reqLogger.Info("got MeteringMultiCloudUI instance, version=" + version)

	// set a default Status value
	if len(instance.Status.PodNames) == 0 {
		instance.Status.PodNames = res.DefaultStatusForCR
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to set MeteringMultiCloudUI default status")
			return reconcile.Result{}, err
		}
	}

	reqLogger.Info("Checking MCM UI Service", "Service.Name", res.McmServiceName)
	// Check if the MCM UI Service already exists, if not create a new one
	newService, err := r.serviceForMCMUI(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = res.ReconcileService(r.client, instance.Namespace, res.McmServiceName, "MCM UI", newService, &needToRequeue)
	if err != nil {
		return reconcile.Result{}, err
	}

	// set common MongoDB env vars based on the instance
	mongoDBEnvVars = res.BuildMongoDBEnvVars(instance.Spec.MongoDB)
	// set common cluster env vars based on the instance
	clusterEnvVars = res.BuildUIClusterEnvVars(instance.Namespace, "", instance.Spec.UI, true)

	// set common Volumes based on the instance
	commonVolumes = res.BuildCommonVolumes(instance.Spec.MongoDB, res.McmDeploymentName, "log4js")

	reqLogger.Info("Checking MCM UI Deployment", "Deployment.Name", res.McmDeploymentName)
	// Check if the MCM UI Deployment already exists, if not create a new one
	newDeployment, err := r.deploymentForMCMUI(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = res.ReconcileDeployment(r.client, instance.Namespace, res.McmDeploymentName, "MCM UI", newDeployment, &needToRequeue)
	if err != nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("Checking MCM UI Ingress", "Ingress.Name", res.McmIngressData.Name)
	// Check if the Ingress already exists, if not create a new one
	newIngress := res.BuildIngress(instance.Namespace, res.McmIngressData)
	// Set MeteringMultiCloudUI instance as the owner and controller of the Ingress
	err = controllerutil.SetControllerReference(instance, newIngress, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for MCM UI Ingress", "Ingress.Namespace", newIngress.Namespace,
			"Ingress.Name", newIngress.Name)
		return reconcile.Result{}, err
	}
	err = res.ReconcileIngress(r.client, instance.Namespace, res.McmIngressData.Name, "MCM UI", newIngress, &needToRequeue)
	if err != nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("Checking MCM UI Certificates")
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

	reqLogger.Info("Updating MeteringMultiCloudUI status")
	// Update the MeteringMultiCloudUI status with the pod names.
	// List the pods for this instance's deployment.
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Namespace),
		client.MatchingLabels(res.LabelsForSelector(res.McmDeploymentName, meteringMcmUICrType, instance.Name)),
	}
	if err = r.client.List(context.TODO(), podList, listOpts...); err != nil {
		reqLogger.Error(err, "Failed to list pods", "MeteringMultiCloudUI.Namespace", instance.Namespace, "MeteringMultiCloudUI.Name", res.McmDeploymentName)
		return reconcile.Result{}, err
	}
	podNames := res.GetPodNames(podList.Items)
	// if no pods were found set the default status
	if len(podNames) == 0 {
		podNames = res.DefaultStatusForCR
	}

	// Update status.PodNames if needed
	if !reflect.DeepEqual(podNames, instance.Status.PodNames) {
		instance.Status.PodNames = podNames
		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update MeteringMultiCloudUI status")
			return reconcile.Result{}, err
		}
	}

	reqLogger.Info("Reconciliation completed")
	// since we updated the status in the CR, sleep 5 seconds to allow the CR to be refreshed.
	time.Sleep(5 * time.Second)
	return reconcile.Result{}, nil
}

// deploymentForMCMUI returns an MCM UI Deployment object
func (r *ReconcileMeteringMultiCloudUI) deploymentForMCMUI(instance *operatorv1alpha1.MeteringMultiCloudUI) (*appsv1.Deployment, error) {
	reqLogger := log.WithValues("func", "deploymentForMCMUI", "instance.Name", instance.Name)
	metaLabels := res.LabelsForMetadata(res.McmDeploymentName)
	selectorLabels := res.LabelsForSelector(res.McmDeploymentName, meteringMcmUICrType, instance.Name)
	podLabels := res.LabelsForPodMetadata(res.McmDeploymentName, meteringMcmUICrType, instance.Name)

	// the InitContainer code is part of the metering-data-manager image
	initImage := res.GetImageID(instance.Spec.ImageRegistry, instance.Spec.ImageTagPostfix,
		res.DefaultImageRegistry, res.DefaultDmImageName, res.VarImageSHAforDM, res.DefaultDmImageTag)
	reqLogger.Info("initImage=" + initImage)
	mcmImage := res.GetImageID(instance.Spec.ImageRegistry, instance.Spec.ImageTagPostfix,
		res.DefaultImageRegistry, res.DefaultMcmUIImageName, res.VarImageSHAforMCMUI, res.DefaultMcmUIImageTag)
	reqLogger.Info("mcmImage=" + mcmImage)

	// The apikey and OIDC secret names can be set in the CR, but will be ignored.
	// We are hardcoding here since the operand is using hardcoded names.
	//OLD if instance.Spec.UI.APIkeySecret != "" {
	//OLD	 apiKeySecretName = instance.Spec.UI.APIkeySecret
	//OLD }
	apiKeySecretName := res.DefaultAPIKeySecretName
	//OLD if instance.Spec.UI.PlatformOidcSecret != "" {
	//OLD	 platformOidcSecretName = instance.Spec.UI.PlatformOidcSecret
	//OLD}
	platformOidcSecretName := res.DefaultPlatformOidcSecretName

	var additionalInfo res.SecretCheckData
	// add to the SECRET_LIST env var
	additionalInfo.Names = apiKeySecretName + " " + platformOidcSecretName + " " + res.McmUICertSecretName
	// add to the SECRET_DIR_LIST env var
	additionalInfo.Dirs = apiKeySecretName + " " + platformOidcSecretName + " " + res.McmUICertDirName
	// add the volume mounts for the secrets
	additionalInfo.VolumeMounts = res.BuildUISecretVolumeMounts(apiKeySecretName, platformOidcSecretName)
	// add the volume mount for the MCM UI cert
	additionalInfo.VolumeMounts = append(additionalInfo.VolumeMounts, res.McmUICertVolumeMountForSecretCheck)

	mcmSecretCheckContainer := res.BuildSecretCheckContainer(res.McmDeploymentName, initImage,
		res.SecretCheckCmd, instance.Spec.MongoDB, &additionalInfo)

	initEnvVars := []corev1.EnvVar{}
	initEnvVars = append(initEnvVars, res.CommonEnvVars...)
	initEnvVars = append(initEnvVars, mongoDBEnvVars...)
	mcmInitContainer := res.BuildInitContainer(res.McmDeploymentName, initImage, initEnvVars)

	mcmMainContainer := res.McmUIMainContainer
	mcmMainContainer.Image = mcmImage
	mcmMainContainer.Name = res.McmDeploymentName
	mcmMainContainer.Env = append(mcmMainContainer.Env, res.IAMEnvVars...)
	mcmMainContainer.Env = append(mcmMainContainer.Env, res.UIEnvVars...)
	mcmMainContainer.Env = append(mcmMainContainer.Env, clusterEnvVars...)
	mcmMainContainer.Env = append(mcmMainContainer.Env, res.CommonEnvVars...)
	mcmMainContainer.Env = append(mcmMainContainer.Env, mongoDBEnvVars...)
	mcmMainContainer.VolumeMounts = append(mcmMainContainer.VolumeMounts, res.CommonMainVolumeMounts...)
	mcmMainContainer.VolumeMounts = append(mcmMainContainer.VolumeMounts, res.McmUICertVolumeMountForMain)

	secretVolumes := res.BuildUISecretVolumes(apiKeySecretName, platformOidcSecretName)
	mcmVolumes := append(commonVolumes, res.McmUICertVolume)
	mcmVolumes = append(mcmVolumes, secretVolumes...)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.McmDeploymentName,
			Namespace: instance.Namespace,
			Labels:    metaLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &instance.Spec.Replicas,
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
	// Set MeteringMultiCloudUI instance as the owner and controller of the Deployment
	err := controllerutil.SetControllerReference(instance, deployment, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for MCM UI Deployment")
		return nil, err
	}
	return deployment, nil
}

// serviceForMCMUI returns an MCM UI Service object
func (r *ReconcileMeteringMultiCloudUI) serviceForMCMUI(instance *operatorv1alpha1.MeteringMultiCloudUI) (*corev1.Service, error) {
	reqLogger := log.WithValues("func", "serviceForMCMUI", "instance.Name", instance.Name)
	metaLabels := res.LabelsForMetadata(res.McmDeploymentName)
	selectorLabels := res.LabelsForSelector(res.McmDeploymentName, meteringMcmUICrType, instance.Name)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.McmServiceName,
			Namespace: instance.Namespace,
			Labels:    metaLabels,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:     "metering-mcm-dashboard",
					Protocol: corev1.ProtocolTCP,
					Port:     3001,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 3001,
					},
				},
			},
			Selector: selectorLabels,
		},
	}

	// Set MeteringMultiCloudUI instance as the owner and controller of the Service
	err := controllerutil.SetControllerReference(instance, service, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for MCM UI Service")
		return nil, err
	}
	return service, nil
}

// Check if the Certificate already exists, if not create a new one.
// This function was created to reduce the cyclomatic complexity :)
func (r *ReconcileMeteringMultiCloudUI) reconcileAllCertificates(instance *operatorv1alpha1.MeteringMultiCloudUI,
	needToRequeue *bool) error {
	reqLogger := log.WithValues("func", "reconcileAllCertificates")

	certificateList := []res.CertificateData{
		res.McmUICertificateData,
	}

	for _, certData := range certificateList {
		reqLogger.Info("Checking Certificate", "Certificate.Name", certData.Name)
		newCertificate := res.BuildCertificate(instance.Namespace, "", certData)
		// Set MeteringMultiCloudUI instance as the owner and controller of the Certificate
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
