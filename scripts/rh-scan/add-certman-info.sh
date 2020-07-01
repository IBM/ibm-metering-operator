#!/bin/bash
#
# Copyright 2020 IBM Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# add certmanager info to the CSV so that the metadata bundle will pass the Red Hat operator scan.
# run as 'scripts/rh-scan/add-certman-info.sh <CSV version>'
# ***** This info is only needed for the scan. It should NOT be in the CSV in the git repo! *****

SED="sed"
unamestr=$(uname)
if [[ "$unamestr" == "Darwin" ]] ; then
    SED=gsed
    type $SED >/dev/null 2>&1 || {
        echo >&2 "$SED it's not installed. Try: brew install gnu-sed" ;
        exit 1;
    }
fi

# check the input parm
CSV_VERSION=$1
if [[ $CSV_VERSION == "" ]]
then
   echo "Missing parm. Need CSV version"
   exit 1
fi

CSV_FILE=bundle/ibm-metering-operator.v${CSV_VERSION}.clusterserviceversion.yaml

# insert the cert-manager lines
LINE1="\    required:"
LINE2="\    - name: certificates.certmanager.k8s.io"
LINE3="\      kind: Certificate"
LINE4="\      version: v1alpha1"
LINE5="\      displayName: Certificate"
LINE6="\      description: Represents a certificate"

$SED -i "/customresourcedefinitions:/a $LINE1\n$LINE2\n$LINE3\n$LINE4\n$LINE5\n$LINE6" "$CSV_FILE"
