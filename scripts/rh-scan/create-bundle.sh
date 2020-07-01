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

###############################################################################
#
# ***** run "make bundle" instead of running this script *****
#
###############################################################################

# create zip file containing the bundle to submit for Red Hat certification
# the bundle consists of package.yaml, clusterserviceversion.yaml, crd.yaml
# run as 'scripts/rh-scan/create-bundle.sh <CSV version>'

# check the input parm
CSV_VERSION=$1
if [[ $CSV_VERSION == "" ]]
then
   echo "Missing parm. Need CSV version"
   exit 1
fi

if [ -d "./bundle" ] 
then
    echo "cleanup bundle directory" 
    rm bundle/*.yaml
    rm bundle/*.zip
else
    echo "create bundle directory"
    mkdir bundle
fi

cp -p deploy/olm-catalog/ibm-metering-operator/ibm-metering-operator.package.yaml bundle/
cp -p deploy/olm-catalog/ibm-metering-operator/$CSV_VERSION/*yaml bundle/
# need certificate-crd.yaml in the bundle so that the operator can be started during the RH scan
cp -p scripts/rh-scan/certificate-crd.yaml bundle/

echo Add certmanager info to CSV for scan
scripts/rh-scan/add-certman-info.sh $CSV_VERSION

cd bundle || exit
zip ibm-metering-metadata ./*.yaml
cd ..
