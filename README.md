# ibm-metering-operator

Operator used to manage the IBM metering service.  The IBM metering service captures detailed usage metrics for your applications and cluster.  The metric data is used for IBM Product licensing compliance and can be viewed in the metering UI or downloaded in a report for use in internal charge-back for cluster workloads.  

## Supported platforms

### Platforms

- OCP 3.11
- OCP 4.1
- OCP 4.2
- OCP 4.3

### Operating Systems

- Linux amd64
- RHEL amd64
- RHEL ppc64le
- RHEL s390x

## Operator versions

- 3.5.0

## Prerequisites

1. Kubernetes 1.11 must be installed
1. OpenShift 3.11+ must be installed
1. IBM Mongodb service - See [IBM mongodb operator](https://github.com/IBM/ibm-mongodb-operator)
1. IBM cert-manager service - See [IBM cert-manager operator](https://github.com/IBM/ibm-cert-manager-operator)
1. IBM IAM service - See [IBM IAM operator](https://github.com/IBM/ibm-iam-operator) </br>Note that this is a soft dependency.  Metering will function without IAM for data collection, but the user interface and report download is not possible.

## Documentation

For installation and configuration, see [IBM Knowledge Center link].

### Developer guide

Information about building and testing the operator.
- Dev quick start
  1. Follow the [ODLM guide](https://github.com/IBM/operand-deployment-lifecycle-manager/blob/master/docs/install/common-service-integration.md#end-to-end-test)
  
- Debugging the operator
  1. Check the Metering or MeteringUI custom resources (CR)

    ````
    kubectl get metering
    kubectl describe metering <metering CR name>
    ````

  1. Look at the logs of the metering-operator pod for errors

    ````
    kubectl get po -n <namespace>
    kubectl logs -n <namespace> <metering-operator pod name>
    ````
    
