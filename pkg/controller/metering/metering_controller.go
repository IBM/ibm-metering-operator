/*
 * Copyright 2020 IBM Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

//CS??? use labelsForMeteringSelect to buld labelsForMeteringPod
//CS??? need to create icp-metering-receiver-secret; see metering-receiver-certificate.yaml
package metering

import (
	"context"
	"reflect"

	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/apimachinery/pkg/api/resource"

	operatorv1alpha1 "github.com/cs-operators/metering-operator/pkg/apis/operator/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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

const dataManagerDeploymentName = "metering-dm"
const readerDaemonSetName = "metering-reader"

var trueVar bool = true
var falseVar bool = false
var defaultMode int32 = 420
var seconds60 int64 = 60
var user99 int64 = 99
var nodeSelector = map[string]string{"management": "true"}
var cpu100 = resource.NewMilliQuantity(100, resource.DecimalSI)          // 100m
var cpu1000 = resource.NewMilliQuantity(1000, resource.DecimalSI)        // 1000m
var memory100 = resource.NewQuantity(100*1024*1024, resource.BinarySI)   // 100Mi
var memory256 = resource.NewQuantity(256*1024*1024, resource.BinarySI)   // 256Mi
var memory2560 = resource.NewQuantity(2560*1024*1024, resource.BinarySI) // 2560Mi
var secretCheckCmd = `set -- $SECRET_LIST; ` +
	`for secretDirName in $SECRET_DIR_LIST; do` +
	`  while true; do` +
	`    echo ` + "`date`" + `: Checking for secret $1;` +
	`    ls /sec/$secretDirName/* && break;` +
	`    echo ` + "`date`" + `: Required secret $1 not found ... try again in 30s;` +
	`    sleep 30;` +
	`  done;` +
	`  echo ` + "`date`" + `: Secret $1 found;` +
	`  shift; ` +
	`done; ` +
	`echo ` + "`date`" + `: All required secrets exist`

var commonEnvVars = []corev1.EnvVar{
	corev1.EnvVar{
		Name:  "NODE_TLS_REJECT_UNAUTHORIZED",
		Value: "0",
	},
	corev1.EnvVar{
		Name: "ICP_API_KEY",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "icp-serviceid-apikey-secret",
				},
				Key:      "ICP_API_KEY",
				Optional: &trueVar,
			},
		},
	},
	corev1.EnvVar{
		Name:  "HC_MONGO_HOST",
		Value: "mongodb",
	},
	corev1.EnvVar{
		Name:  "HC_MONGO_PORT",
		Value: "27017",
	},
	corev1.EnvVar{
		Name: "HC_MONGO_USER",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "icp-mongodb-admin",
				},
				Key:      "user",
				Optional: &trueVar,
			},
		},
	},
	corev1.EnvVar{
		Name: "HC_MONGO_PASS",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "icp-mongodb-admin",
				},
				Key:      "password",
				Optional: &trueVar,
			},
		},
	},
	corev1.EnvVar{
		Name:  "HC_MONGO_ISSSL",
		Value: "true",
	},
	corev1.EnvVar{
		Name:  "HC_MONGO_SSL_CA",
		Value: "/certs/mongodb-ca/tls.crt",
	},
	corev1.EnvVar{
		Name:  "HC_MONGO_SSL_CERT",
		Value: "/certs/mongodb-client/tls.crt",
	},
	corev1.EnvVar{
		Name:  "HC_MONGO_SSL_KEY",
		Value: "/certs/mongodb-client/tls.key",
	},
}

var dmSecretCheckContainer = corev1.Container{
	Image:           "temp",
	Name:            "metering-dm-secret-check",
	ImagePullPolicy: corev1.PullAlways,
	Command: []string{
		"sh",
		"-c",
		secretCheckCmd,
	},
	Env: []corev1.EnvVar{
		corev1.EnvVar{
			Name:  "SECRET_LIST",
			Value: "icp-serviceid-apikey-secret icp-mongodb-admin icp-mongodb-admin cluster-ca-cert icp-mongodb-client-cert",
			//CS??? Value: "icp-serviceid-apikey-secret icp-metering-receiver-secret icp-mongodb-admin icp-mongodb-admin cluster-ca-cert icp-mongodb-client-cert",
		},
		corev1.EnvVar{
			Name:  "SECRET_DIR_LIST",
			Value: "icp-serviceid-apikey-secret muser-icp-mongodb-admin mpass-icp-mongodb-admin cluster-ca-cert icp-mongodb-client-cert",
			//CS??? Value: "icp-serviceid-apikey-secret icp-metering-receiver-secret muser-icp-mongodb-admin mpass-icp-mongodb-admin cluster-ca-cert icp-mongodb-client-cert",
		},
	},
	VolumeMounts: []corev1.VolumeMount{
		corev1.VolumeMount{
			Name:      "icp-serviceid-apikey-secret",
			MountPath: "/sec/icp-serviceid-apikey-secret",
		},
		corev1.VolumeMount{
			Name:      "icp-metering-receiver-certs",
			MountPath: "/sec/icp-metering-receiver-secret",
		},
		corev1.VolumeMount{
			Name:      "mongodb-ca-cert",
			MountPath: "/sec/cluster-ca-cert",
		},
		corev1.VolumeMount{
			Name:      "mongodb-client-cert",
			MountPath: "/sec/icp-mongodb-client-cert",
		},
		corev1.VolumeMount{
			Name:      "muser-icp-mongodb-admin",
			MountPath: "/sec/muser-icp-mongodb-admin",
		},
		corev1.VolumeMount{
			Name:      "mpass-icp-mongodb-admin",
			MountPath: "/sec/mpass-icp-mongodb-admin",
		},
	},
	Resources: corev1.ResourceRequirements{
		Limits: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    *cpu100,
			corev1.ResourceMemory: *memory100},
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    *cpu100,
			corev1.ResourceMemory: *memory100},
	},
	SecurityContext: &corev1.SecurityContext{
		AllowPrivilegeEscalation: &falseVar,
		Privileged:               &falseVar,
		ReadOnlyRootFilesystem:   &trueVar,
		RunAsNonRoot:             &trueVar,
		RunAsUser:                &user99,
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{
				"ALL",
			},
		},
	},
}
var dmInitContainer = corev1.Container{
	Image:           "temp",
	Name:            "metering-dm-init",
	ImagePullPolicy: corev1.PullAlways,
	Command: []string{
		"node",
		"/datamanager/lib/metering_init.js",
		"verifyOnlyMongo",
	},
	Env: []corev1.EnvVar{
		corev1.EnvVar{
			Name:  "MCM_VERBOSE",
			Value: "true",
		},
	},
	VolumeMounts: []corev1.VolumeMount{
		corev1.VolumeMount{
			Name:      "mongodb-ca-cert",
			MountPath: "/certs/mongodb-ca",
		},
		corev1.VolumeMount{
			Name:      "mongodb-client-cert",
			MountPath: "/certs/mongodb-client",
		},
	},
	Resources: corev1.ResourceRequirements{
		Limits: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    *cpu100,
			corev1.ResourceMemory: *memory100},
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    *cpu100,
			corev1.ResourceMemory: *memory100},
	},
	SecurityContext: &corev1.SecurityContext{
		AllowPrivilegeEscalation: &falseVar,
		Privileged:               &falseVar,
		ReadOnlyRootFilesystem:   &trueVar,
		RunAsNonRoot:             &trueVar,
		RunAsUser:                &user99,
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{
				"ALL",
			},
		},
	},
}
var dmMainContainer = corev1.Container{
	Image: "temp",
	//CS??? Image: "hyc-cloud-private-edge-docker-local.artifactory.swg-devops.com/ibmcom-amd64/metering-data-manager:3.3.1",
	Name:            dataManagerDeploymentName,
	ImagePullPolicy: corev1.PullAlways,
	VolumeMounts: []corev1.VolumeMount{
		corev1.VolumeMount{
			Name:      "mongodb-ca-cert",
			MountPath: "/certs/mongodb-ca",
		},
		corev1.VolumeMount{
			Name:      "mongodb-client-cert",
			MountPath: "/certs/mongodb-client",
		},
		corev1.VolumeMount{
			Name:      "icp-metering-receiver-certs",
			MountPath: "/certs/metering-receiver",
		},
	},
	Env: []corev1.EnvVar{
		corev1.EnvVar{
			Name:  "METERING_API_ENABLED",
			Value: "true",
		},
		corev1.EnvVar{
			Name:  "HC_DM_USE_HTTPS",
			Value: "false",
		},
		corev1.EnvVar{
			Name:  "HC_DM_MCM_RECEIVER_ENABLED",
			Value: "true",
		},
		corev1.EnvVar{
			Name:  "HC_DM_MCM_SENDER_ENABLED",
			Value: "false",
		},
		corev1.EnvVar{
			Name:  "HC_DM_STORAGEREADER_ENABLED",
			Value: "true",
		},
		corev1.EnvVar{
			Name:  "HC_DM_REPORTER2_ENABLED",
			Value: "true",
		},
		corev1.EnvVar{
			Name:  "HC_DM_PURGER2_ENABLED",
			Value: "true",
		},
		corev1.EnvVar{
			Name:  "HC_DM_PREAGGREGATOR_ENABLED",
			Value: "true",
		},
		corev1.EnvVar{
			Name:  "HC_DM_METRICS_ENABLED",
			Value: "true",
		},
		corev1.EnvVar{
			Name:  "HC_DM_SELFMETER_PURGER_ENABLED",
			Value: "true",
		},
		corev1.EnvVar{
			Name:  "HC_DM_IS_ICP",
			Value: "true",
		},
		corev1.EnvVar{
			Name:  "HC_RECEIVER_SSL_CA",
			Value: "/certs/metering-receiver/ca.crt",
		},
		corev1.EnvVar{
			Name:  "HC_RECEIVER_SSL_CERT",
			Value: "/certs/metering-receiver/tls.crt",
		},
		corev1.EnvVar{
			Name:  "HC_RECEIVER_SSL_KEY",
			Value: "/certs/metering-receiver/tls.key",
		},
		corev1.EnvVar{
			Name:  "CLUSTER_NAME",
			Value: "mycluster",
		},
		corev1.EnvVar{
			Name:  "DEFAULT_IAM_TOKEN_SERVICE_PORT",
			Value: "10443",
		},
		corev1.EnvVar{
			Name:  "DEFAULT_IAM_PAP_SERVICE_PORT",
			Value: "39001",
		},
		corev1.EnvVar{
			Name:  "HC_DM_ALLOW_TEST",
			Value: "false",
		},
	},
	Ports: []corev1.ContainerPort{
		{ContainerPort: 3000},
		{ContainerPort: 5000},
	},
	LivenessProbe: &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/livenessProbe",
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 3000,
				},
				Scheme: "",
			},
		},
		InitialDelaySeconds: 305,
		TimeoutSeconds:      5,
		PeriodSeconds:       300,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	},
	ReadinessProbe: &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/readinessProbe",
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 3000,
				},
				Scheme: "",
			},
		},
		InitialDelaySeconds: 15,
		TimeoutSeconds:      15,
		PeriodSeconds:       30,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	},
	Resources: corev1.ResourceRequirements{
		Limits: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    *cpu1000,
			corev1.ResourceMemory: *memory2560},
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    *cpu100,
			corev1.ResourceMemory: *memory256},
	},
	SecurityContext: &corev1.SecurityContext{
		AllowPrivilegeEscalation: &falseVar,
		Privileged:               &falseVar,
		ReadOnlyRootFilesystem:   &trueVar,
		RunAsNonRoot:             &trueVar,
		RunAsUser:                &user99,
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{
				"ALL",
			},
		},
	},
}

var log = logf.Log.WithName("controller_metering")

type MeteringDeploymentSpec struct {
	Enabled   bool
	ImageRepo string
	ImageTag  string
}

var dmDeployment = MeteringDeploymentSpec{
	Enabled:   true,
	ImageRepo: "CS???",
	ImageTag:  "CS???",
}

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

	// Fetch the Metering instance
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
		reqLogger.Error(err, "Failed to get Metering")
		return reconcile.Result{}, err
	}

	opVersion := instance.Spec.OperatorVersion
	reqLogger.Info("CS??? got Metering instance, version=" + opVersion + ", checking Service")
	// Check if the service already exists, if not create a new one
	currentService := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: dataManagerDeploymentName, Namespace: instance.Namespace}, currentService)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		newService := r.serviceForDataMgr(instance)
		reqLogger.Info("Creating a new Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
		err = r.client.Create(context.TODO(), newService)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
			return reconcile.Result{}, err
		}
		// Service created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Service")
		return reconcile.Result{}, err
	}

	reqLogger.Info("CS??? got Service, checking Deployment")
	// Check if the deployment already exists, if not create a new one
	currentDeployment := &appsv1.Deployment{}
	//CS??? err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, currentDeployment)
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: dataManagerDeploymentName, Namespace: instance.Namespace}, currentDeployment)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		newDeployment := r.deploymentForDataMgr(instance)
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", newDeployment.Namespace, "Deployment.Name", newDeployment.Name)
		err = r.client.Create(context.TODO(), newDeployment)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", newDeployment.Namespace, "Deployment.Name", newDeployment.Name)
			return reconcile.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Deployment")
		return reconcile.Result{}, err
	}

	reqLogger.Info("CS??? got Deployment, checking DaemonSet")
	// Check if the DaemonSet already exists, if not create a new one
	currentDaemonSet := &appsv1.DaemonSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: readerDaemonSetName, Namespace: instance.Namespace}, currentDaemonSet)
	if err != nil && errors.IsNotFound(err) {
		// Define a new DaemonSet
		newDaemonSet := r.daemonForReader(instance)
		reqLogger.Info("Creating a new DaemonSet", "Deployment.Namespace", newDaemonSet.Namespace, "Deployment.Name", newDaemonSet.Name)
		err = r.client.Create(context.TODO(), newDaemonSet)
		if err != nil {
			reqLogger.Error(err, "Failed to create new DaemonSet", "Deployment.Namespace", newDaemonSet.Namespace, "Deployment.Name", newDaemonSet.Name)
			return reconcile.Result{}, err
		}
		// DaemonSet created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get DaemonSet")
		return reconcile.Result{}, err
	}

	reqLogger.Info("CS??? checking current deployment")
	// Ensure the deployment size is the same as the spec
	size := instance.Spec.DataManager.Size
	if *currentDeployment.Spec.Replicas != size {
		currentDeployment.Spec.Replicas = &size
		reqLogger.Info("CS??? updating current deployment")
		err = r.client.Update(context.TODO(), currentDeployment)
		if err != nil {
			reqLogger.Error(err, "Failed to update Deployment", "Deployment.Namespace", currentDeployment.Namespace, "Deployment.Name", currentDeployment.Name)
			return reconcile.Result{}, err
		}
		// Spec updated - return and requeue
		return reconcile.Result{Requeue: true}, nil
	}

	//CS??? reqLogger.Info("CS??? checking current DaemonSet")

	reqLogger.Info("CS??? updating Metering status")
	// Update the Metering status with the pod names
	// List the pods for this instance's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Namespace),
		client.MatchingLabels(labelsForMeteringSelect(instance.Name, dataManagerDeploymentName)),
	}
	if err = r.client.List(context.TODO(), podList, listOpts...); err != nil {
		reqLogger.Error(err, "Failed to list pods", "Metering.Namespace", instance.Namespace, "Metering.Name", dataManagerDeploymentName)
		return reconcile.Result{}, err
	}
	reqLogger.Info("CS??? get pod names")
	podNames := getPodNames(podList.Items)

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

// deploymentForDataMgr returns a DataManager Deployment object
func (r *ReconcileMetering) deploymentForDataMgr(instance *operatorv1alpha1.Metering) *appsv1.Deployment {
	reqLogger := log.WithValues("deploymentForDataMgr", "Entry", "instance.Name", instance.Name)
	labels1 := labelsForMeteringPod(instance.Name, dataManagerDeploymentName)
	labels2 := labelsForMeteringSelect(instance.Name, dataManagerDeploymentName)
	labels3 := labelsForMeteringMeta(dataManagerDeploymentName)

	replicas := instance.Spec.DataManager.Size
	image := instance.Spec.DataManager.ImageRepo + ":" + instance.Spec.DataManager.ImageTag
	reqLogger.Info("CS??? image=" + image)

	dmSecretCheckContainer.Image = image
	dmInitContainer.Image = image
	dmMainContainer.Image = image
	dmInitContainer.Env = append(dmInitContainer.Env, commonEnvVars...)
	dmMainContainer.Env = append(dmMainContainer.Env, commonEnvVars...)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dataManagerDeploymentName,
			Namespace: instance.Namespace,
			Labels:    labels3,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels2,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels1,
				},
				Spec: corev1.PodSpec{
					NodeSelector:                  nodeSelector,
					PriorityClassName:             "system-cluster-critical",
					TerminationGracePeriodSeconds: &seconds60,
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									corev1.NodeSelectorTerm{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											corev1.NodeSelectorRequirement{
												Key:      "beta.kubernetes.io/arch",
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{"amd64"},
											},
										},
									},
								},
							},
						},
					},
					Tolerations: []corev1.Toleration{
						corev1.Toleration{
							Key:      "dedicated",
							Operator: corev1.TolerationOpExists,
							Effect:   corev1.TaintEffectNoSchedule,
						},
						corev1.Toleration{
							Key:      "CriticalAddonsOnly",
							Operator: corev1.TolerationOpExists,
						},
					},
					Volumes: []corev1.Volume{
						corev1.Volume{
							Name: "mongodb-ca-cert",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  "cluster-ca-cert",
									DefaultMode: &defaultMode,
									Optional:    &trueVar,
								},
							},
						},
						corev1.Volume{
							Name: "mongodb-client-cert",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  "icp-mongodb-client-cert",
									DefaultMode: &defaultMode,
									Optional:    &trueVar,
								},
							},
						},
						corev1.Volume{
							Name: "muser-icp-mongodb-admin",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  "icp-mongodb-admin",
									DefaultMode: &defaultMode,
									Optional:    &trueVar,
								},
							},
						},
						corev1.Volume{
							Name: "mpass-icp-mongodb-admin",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  "icp-mongodb-admin",
									DefaultMode: &defaultMode,
									Optional:    &trueVar,
								},
							},
						},
						corev1.Volume{
							Name: "icp-metering-receiver-certs",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  "icp-metering-receiver-secret",
									DefaultMode: &defaultMode,
									Optional:    &trueVar,
								},
							},
						},
						corev1.Volume{
							Name: "icp-serviceid-apikey-secret",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  "icp-serviceid-apikey-secret",
									DefaultMode: &defaultMode,
									Optional:    &trueVar,
								},
							},
						},
					},
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
	controllerutil.SetControllerReference(instance, deployment, r.scheme)
	return deployment
}

// serviceForDataMgr returns a DataManager Service object
func (r *ReconcileMetering) serviceForDataMgr(instance *operatorv1alpha1.Metering) *corev1.Service {
	reqLogger := log.WithValues("serviceForDataMgr", "Entry", "instance.Name", instance.Name)
	labels1 := labelsForMeteringSelect(instance.Name, dataManagerDeploymentName)
	labels2 := labelsForMeteringMeta(dataManagerDeploymentName)

	reqLogger.Info("CS??? Entry")
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        dataManagerDeploymentName,
			Namespace:   instance.Namespace,
			Labels:      labels2,
			Annotations: map[string]string{"prometheus.io/scrape": "false", "prometheus.io/scheme": "http"},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name: "datamanager",
					Port: 3000},
			},
			Selector: labels1,
		},
	}

	// Set Metering instance as the owner and controller of the Service
	controllerutil.SetControllerReference(instance, service, r.scheme)
	return service
}

// daemonForReader returns a Reader DaemonSet object
func (r *ReconcileMetering) daemonForReader(instance *operatorv1alpha1.Metering) *appsv1.DaemonSet {
	reqLogger := log.WithValues("serviceForDataMgr", "Entry", "instance.Name", instance.Name)
	//CS??? labels1 := labelsForMeteringSelect(instance.Name, readerDaemonSetName)
	labels2 := labelsForMeteringMeta(readerDaemonSetName)

	reqLogger.Info("CS??? Entry")
	daemon := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      readerDaemonSetName,
			Namespace: instance.Namespace,
			Labels:    labels2,
		},
		Spec: appsv1.DaemonSetSpec{},
	}

	// Set Metering instance as the owner and controller of the DaemonSet
	controllerutil.SetControllerReference(instance, daemon, r.scheme)
	return daemon
}

// labelsForMetering returns the labels for selecting the resources
// belonging to the given metering CR name.
//CS??? need separate func for each image to set "instanceName"???
func labelsForMeteringPod(instanceName string, deploymentName string) map[string]string {
	return map[string]string{"app": deploymentName, "component": "meteringsvc", "metering_cr": instanceName, "app.kubernetes.io/name": deploymentName, "app.kubernetes.io/component": "meteringsvc", "release": "metering"}
	//CS??? return map[string]string{"app": deploymentName, "component": "meteringsvc", "metering_cr": instanceName}
	//CS??? return map[string]string{"app.kubernetes.io/name": deploymentName, "app.kubernetes.io/component": "meteringsvc", "metering_cr": instanceName}
}

//CS??? need separate func for each image to set "app"???
func labelsForMeteringSelect(instanceName string, deploymentName string) map[string]string {
	return map[string]string{"app": deploymentName, "component": "meteringsvc", "metering_cr": instanceName}
}

//CS???
func labelsForMeteringMeta(deploymentName string) map[string]string {
	return map[string]string{"app.kubernetes.io/name": deploymentName, "app.kubernetes.io/component": "meteringsvc", "release": "metering"}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	reqLogger := log.WithValues("Request.Namespace", "CS??? namespace", "Request.Name", "CS???")
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
		reqLogger.Info("CS??? pod name=" + pod.Name)
	}
	return podNames
}
