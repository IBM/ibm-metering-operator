# ibm-metering-operator

> **Important:** Do not install this operator directly. Only install this operator using the IBM Common Services Operator.
> For more information about installing this operator and other Common Services operators, see [Installer documentation](http://ibm.biz/cpcs_opinstall).
> If you are using this operator as part of an IBM Cloud Pak, see the documentation for that IBM Cloud Pak to learn more about how to install and use the operator service.
> For more information about IBM Cloud Paks, see [IBM Cloud Paks that use Common Services](http://ibm.biz/cpcs_cloudpaks).

You can use the ibm-metering-operator to install the Metering service for the IBM Cloud Platform Common Services.
The Metering service captures detailed usage metrics for your applications and cluster. The metric data is used for IBM Product licensing compliance and can be viewed in the metering UI or downloaded in a report for use in internal charge-back for cluster workloads.

## Supported platforms

Red Hat OpenShift Container Platform 4.3 or newer installed on one of the following platforms:
   - Linux x86_64
   - Linux on Power (ppc64le)
   - Linux on IBM Z and LinuxONE

## Operator versions

- 3.6.0
- 3.5.0

## Prerequisites

The Metering service has dependencies on other IBM Cloud Platform Common Services.
Before you install this operator, you need to first install the operator dependencies and prerequisites:
- For the list of operator dependencies, see the IBM Knowledge Center [Common Services dependencies documentation](http://ibm.biz/cpcs_opdependencies).
- For the list of prerequisites for installing the operator, see the IBM Knowledge Center [Preparing to install services documentation](http://ibm.biz/cpcs_opinstprereq).
- **Note:** The Metering service has a soft dependency on `ibm-iam-operator` and `ibm-commonui-operator`. Metering functions without those operators for data collection, but the user interface and report download are not possible.

## Documentation

To install the operator with the IBM Common Services Operator follow the the installation and configuration instructions within the IBM Knowledge Center.
- If you are using the operator as part of an IBM Cloud Pak, see the documentation for that IBM Cloud Pak. For a list of IBM Cloud Paks, see [IBM Cloud Paks that use Common Services](http://ibm.biz/cpcs_cloudpaks).
- If you are using the operator with an IBM Containerized Software, see the IBM Cloud Platform Common Services Knowledge Center [Installer documentation](http://ibm.biz/cpcs_opinstall).

## SecurityContextConstraints Requirements

The Metering service supports running with the OpenShift Container Platform 4.3 default restricted Security Context Constraints (SCCs).

For more information about the OpenShift Container Platform Security Context Constraints, see [Managing Security Context Constraints](https://docs.openshift.com/container-platform/4.3/authentication/managing-security-context-constraints.html).

### OCP 4.3 restricted SCC

```yaml
allowHostDirVolumePlugin: false
allowHostIPC: false
allowHostNetwork: false
allowHostPID: false
allowHostPorts: false
allowPrivilegeEscalation: true
allowPrivilegedContainer: false
allowedCapabilities: null
apiVersion: security.openshift.io/v1
defaultAddCapabilities: null
fsGroup:
  type: MustRunAs
groups:
- system:authenticated
kind: SecurityContextConstraints
metadata:
  annotations:
    kubernetes.io/description: restricted denies access to all host features and requires
      pods to be run with a UID, and SELinux context that are allocated to the namespace.  This
      is the most restrictive SCC and it is used by default for authenticated users.
  creationTimestamp: "2020-03-27T15:01:00Z"
  generation: 1
  name: restricted
  resourceVersion: "6365"
  selfLink: /apis/security.openshift.io/v1/securitycontextconstraints/restricted
  uid: 6a77775c-a6d8-4341-b04c-bd826a67f67e
priority: null
readOnlyRootFilesystem: false
requiredDropCapabilities:
- KILL
- MKNOD
- SETUID
- SETGID
runAsUser:
  type: MustRunAsRange
seLinuxContext:
  type: MustRunAs
supplementalGroups:
  type: RunAsAny
users: []
volumes:
- configMap
- downwardAPI
- emptyDir
- persistentVolumeClaim
- projected
- secret
```

### Developer guide

If, as a developer, you are looking to build and test this operator to try out and learn more about the operator and its capabilities,
you can use the following developer guide. This guide provides commands for a quick install and initial validation for running the operator.

> **Important:** The following developer guide is provided as-is and only for trial and education purposes. IBM and IBM Support does not provide any support for the usage of the operator with this developer guide. For the official supported install and usage guide for the operator, see the the IBM Knowledge Center documentation for your IBM Cloud Pak or for IBM Cloud Platform Common Services.

### Quick start guide

- Follow the [IBM Common Services Operator guide](https://github.com/IBM/ibm-common-service-operator/blob/master/docs/install.md).

- Use the following quick start commands for building and testing the operator:

```bash
oc login -u <CLUSTER_USER> -p <CLUSTER_PASS> <CLUSTER_IP>:6443

export NAMESPACE=ibm-common-services
export BASE_DIR=ibm-metering-operator
export CSV_VERSION=3.6.0

make install
```

- To uninstall the operator installed using `make install`, run `make uninstall`.

### Debugging guide

1. Check the Metering custom resources (CR): `metering`, `meteringui`, `meteringreportserver`

    ````
    kubectl get metering
    kubectl describe metering <metering CR name>
    ````

1. Look at the logs of the ibm-metering-operator pod for errors.

    ````
    kubectl -n ibm-common-services get pods | grep metering
    kubectl -n ibm-common-services logs <ibm-metering-operator pod name>
    ````

1. Look at the logs of the metering operand pods for errors: `metering-dm`, `metering-reader`, `metering-ui`, `metering-report`

    ````
    kubectl -n ibm-common-services get pods | grep metering
    kubectl -n ibm-common-services logs <metering-dm pod name>
    ````

### End-to-End testing

For more instructions on how to run end-to-end testing with the Operand Deployment Lifecycle Manager, see [ODLM guide](https://github.com/IBM/operand-deployment-lifecycle-manager/blob/master/docs/dev/e2e.md).
