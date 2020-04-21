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

package meteringreport

import (
	"context"
	"reflect"
	"time"

	operatorv1alpha1 "github.com/ibm/ibm-metering-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-metering-operator/pkg/resources"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const meteringReportCrType = "meteringreport_cr"

var log = logf.Log.WithName("controller_meteringreport")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new MeteringReport Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	ns, _ := k8sutil.GetWatchNamespace()

	if ns == "" {
		ns = "ibm-common-services"
	}

	return &ReconcileMeteringReport{
		client:         mgr.GetClient(),
		scheme:         mgr.GetScheme(),
		watchNamespace: ns,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("meteringreport-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource MeteringReport
	err = c.Watch(&source.Kind{Type: &operatorv1alpha1.MeteringReport{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner MeteringReport
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.MeteringReport{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileMeteringReport implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileMeteringReport{}

// ReconcileMeteringReport reconciles a MeteringReport object
type ReconcileMeteringReport struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client         client.Client
	scheme         *runtime.Scheme
	watchNamespace string
}

// Reconcile reads that state of the cluster for a MeteringReport object and makes changes based on the state read
// and what is in the MeteringReport.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMeteringReport) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling MeteringReport")

	// if we need to create several resources, set a flag so we just requeue one time instead of after each create.
	needToRequeue := false

	// Fetch the MeteringReport instance
	instance := &operatorv1alpha1.MeteringReport{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	newAPIService, err := r.apiserviceForReport(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = reconcileAPIService(r.client, newAPIService, &needToRequeue)
	if err != nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("Checking Report Deployment", "Deployment.Name", res.ReportDeploymentName)
	// Check if the Report Deployment already exists, if not create a new one
	newReportDeployment, err := r.deploymentForReport(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = res.ReconcileDeployment(r.client, r.watchNamespace, res.ReportDeploymentName, "Report", newReportDeployment, &needToRequeue)
	if err != nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("Checking Report Service", "Service.Name", res.ReportServiceName)
	// Check if the Report Service already exists, if not create a new one
	newReportService, err := r.serviceForReport(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = res.ReconcileService(r.client, r.watchNamespace, res.ReportServiceName, "Report", newReportService, &needToRequeue)
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

	reqLogger.Info("Updating MeteringReport status")
	// Update the MeteringReport status with the pod names.
	// List the pods for this instance's deployment.
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Namespace),
		client.MatchingLabels(res.LabelsForSelector(res.ReportDeploymentName, meteringReportCrType, instance.Name)),
	}
	if err = r.client.List(context.TODO(), podList, listOpts...); err != nil {
		reqLogger.Error(err, "Failed to list pods", "MeteringReport.Namespace", instance.Namespace,
			"MeteringReport.Name", res.ReportDeploymentName)
		return reconcile.Result{}, err
	}
	podNames := res.GetPodNames(podList.Items)

	// Update status.PodNames if needed
	if !reflect.DeepEqual(podNames, instance.Status.PodNames) {
		instance.Status.PodNames = podNames
		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update MeteringReport status")
			return reconcile.Result{}, err
		}
	}

	reqLogger.Info("Reconciliation completed")
	// since we updated the status in the Metering CR, sleep 5 seconds to allow the CR to be refreshed.
	time.Sleep(5 * time.Second)
	return reconcile.Result{}, nil
}

// deploymentForReport returns a MeteringReport Deployment object
func (r *ReconcileMeteringReport) deploymentForReport(instance *operatorv1alpha1.MeteringReport) (*appsv1.Deployment, error) {
	reqLogger := log.WithValues("func", "deploymentForReport", "instance.Name", instance.Name)
	metaLabels := res.LabelsForMetadata(res.ReportDeploymentName)
	selectorLabels := res.LabelsForSelector(res.ReportDeploymentName, meteringReportCrType, instance.Name)
	podLabels := res.LabelsForPodMetadata(res.ReportDeploymentName, meteringReportCrType, instance.Name)

	var reportImage, imageRegistry string
	if instance.Spec.ImageRegistry == "" {
		imageRegistry = res.DefaultImageRegistry
		reqLogger.Info("use default imageRegistry=" + imageRegistry)
	} else {
		imageRegistry = instance.Spec.ImageRegistry
		reqLogger.Info("use instance imageRegistry=" + imageRegistry)
	}
	reportImage = imageRegistry + "/" + res.DefaultReportImageName + ":" + res.DefaultReportImageTag + instance.Spec.ImageTagPostfix
	reqLogger.Info("reportImage=" + reportImage)

	ReportContainer := res.ReportContainer
	ReportContainer.Image = reportImage

	reportVolumes := []corev1.Volume{res.TempDirVolume, res.APICertVolume}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.ReportDeploymentName,
			Namespace: r.watchNamespace,
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
					ServiceAccountName: res.GetServiceAccountName(),
					HostNetwork:        false,
					HostPID:            false,
					HostIPC:            false,
					Volumes:            reportVolumes,
					Containers: []corev1.Container{
						ReportContainer,
					},
				},
			},
		},
	}
	// Set Metering instance as the owner and controller of the Deployment
	err := controllerutil.SetControllerReference(instance, deployment, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for Report Deployment")
		return nil, err
	}
	return deployment, nil
}

// serviceForReport returns a Report Service object
func (r *ReconcileMeteringReport) serviceForReport(instance *operatorv1alpha1.MeteringReport) (*corev1.Service, error) {
	reqLogger := log.WithValues("func", "serviceForReport", "instance.Name", instance.Name)
	metaLabels := res.LabelsForMetadata(res.ReportServiceName)
	selectorLabels := res.LabelsForSelector(res.ReportDeploymentName, meteringReportCrType, instance.Name)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.ReportServiceName,
			Namespace: r.watchNamespace,
			Labels:    metaLabels,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Protocol: corev1.ProtocolTCP,
					Port:     443,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 7443,
					},
				},
			},
			Selector: selectorLabels,
		},
	}

	// Set Metering instance as the owner and controller of the Service
	err := controllerutil.SetControllerReference(instance, service, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for Report Service")
		return nil, err
	}
	return service, nil
}

func (r *ReconcileMeteringReport) apiserviceForReport(instance *operatorv1alpha1.MeteringReport) (*apiregistrationv1.APIService, error) {
	reqLogger := log.WithValues("func", "apiserviceForReport", "instance.Name", instance.Name)
	metaLabels := res.LabelsForMetadata(res.ReportDeploymentName)
	// APIService is cluster-scoped, so don't set Namespace in ObjectMeta.
	// Use the watchNamespace from the operator to set Spec.Service.Namespace
	apiservice := &apiregistrationv1.APIService{
		ObjectMeta: metav1.ObjectMeta{
			Name:   res.DefaultAPIServiceName,
			Labels: metaLabels,
		},
		Spec: apiregistrationv1.APIServiceSpec{
			InsecureSkipTLSVerify: true,
			Version:               "v1",
			Group:                 "metering.ibm.com",
			GroupPriorityMinimum:  1000,
			VersionPriority:       15,
			Service: &apiregistrationv1.ServiceReference{
				Name:      "metering-report",
				Namespace: r.watchNamespace,
			},
		},
	}

	// Since both the APIService and the instance (CR) are cluster-scoped,
	// we can set the MeteringReport instance as the owner and controller of the APIService.
	err := controllerutil.SetControllerReference(instance, apiservice, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for APIService")
		return nil, err
	}
	return apiservice, nil
}

func reconcileAPIService(client client.Client, newAPIService *apiregistrationv1.APIService, needToRequeue *bool) error {
	logger := log.WithValues("func", "ReconcileAPIService")

	currentAPIService := &apiregistrationv1.APIService{}

	// APIService is cluster-scoped, so set Namespace to ""
	err := client.Get(context.TODO(), types.NamespacedName{Name: res.DefaultAPIServiceName, Namespace: ""}, currentAPIService)
	if err != nil && errors.IsNotFound(err) {
		// Create a new APIService
		logger.Info("Creating a new APIService", "APIService.Name", newAPIService.Name)
		err := client.Create(context.TODO(), newAPIService)
		if err != nil && errors.IsAlreadyExists(err) {
			// Already exists from previous reconcile, requeue
			logger.Info("APIService already exists")
			*needToRequeue = true
		} else if err != nil {
			logger.Error(err, "Failed to create new APIService",
				"APIService.Name", newAPIService.Name)
			return err
		} else {
			// Deployment created successfully - return and requeue
			*needToRequeue = true
		}
	} else if err != nil {
		logger.Error(err, "Failed to get APIService", "APIService.Name", newAPIService)
		return err
	} else {
		// Found apiservice, so determine if the resource has changed
		logger.Info("Comparing APIService")
		if !res.IsAPIServiceEqual(currentAPIService, newAPIService) {
			logger.Info("Updating APIService", "APIService.Name", currentAPIService.Name)
			currentAPIService.ObjectMeta.Name = newAPIService.ObjectMeta.Name
			currentAPIService.ObjectMeta.Labels = newAPIService.ObjectMeta.Labels
			currentAPIService.Spec = newAPIService.Spec
			err = client.Update(context.TODO(), currentAPIService)
			if err != nil {
				logger.Error(err, "Failed to update APIService", "APIService.Name", currentAPIService.Name)
				return err
			}
		}
	}
	return nil
}
