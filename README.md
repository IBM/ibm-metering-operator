# ibm-metering-operator

Operator used to manage the IBM metering service. The IBM metering service captures detailed usage metrics for your applications and cluster. The metric data is used for IBM Product licensing compliance and can be viewed in the metering UI or downloaded in a report for use in internal charge-back for cluster workloads.

## Supported platforms

- OCP 3.11
- OCP 4.1
- OCP 4.2
- OCP 4.3

## Operating Systems

- Linux amd64
- RHEL amd64
- RHEL ppc64le
- RHEL s390x

## Operator versions

- 3.5.0

## Prerequisites

1. Kubernetes 1.11 must be installed.
1. OpenShift 3.11+ must be installed.
1. IBM MongoDB service - See [IBM MongoDB operator](https://github.com/IBM/ibm-mongodb-operator).
1. IBM Certificate manager service - See [IBM cert-manager operator](https://github.com/IBM/ibm-cert-manager-operator).
1. IBM IAM service - See [IBM IAM operator](https://github.com/IBM/ibm-iam-operator). </br>**Note:** This service is a soft dependency. Metering functions without IAM for data collection, but the user interface and report download are not possible.

## SecurityContextConstraints Requirements

The metering service supports running under the OpenShift Container Platform default restricted security context constraints.

OCP 4.3 restricted SCC

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

## Documentation

For installation and configuration, see the [IBM Cloud Platform Common Services documentation](http://ibm.biz/cpcsdocs).

### Developer guide

Information about building and testing the operator.
- Developer quick start
  1. Follow the [ODLM guide](https://github.com/IBM/operand-deployment-lifecycle-manager/blob/master/docs/install/common-service-integration.md#end-to-end-test).

- Debugging the operator
  1. Check the metering or metering UI custom resources (CR).

    ````
    kubectl get metering
    kubectl describe metering <metering CR name>
    ````

  1. Look at the logs of the metering-operator pod for errors.

    ````
    kubectl get po -n <namespace>
    kubectl logs -n <namespace> <metering-operator pod name>
    ````
