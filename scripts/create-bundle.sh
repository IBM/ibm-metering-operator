#!/bin/bash

# create zip file containing the bundle to submit for Red Hat certification
# the bundle consists of package.yaml, clusterserviceversion.yaml, crd.yaml
# run as 'scripts/create-bundle.sh'

if [ -d "./bundle" ] 
then
    echo "cleanup bundle directory" 
    rm bundle/*.yaml
    rm bundle/*.zip
else
    echo "create bundle directory"
    mkdir bundle
fi

#ibm-metering-operator.package.yaml
#ibm-metering-operator.v3.5.0.clusterserviceversion.yaml
#operator.ibm.com_meteringmulticlouduis_crd.yaml
#operator.ibm.com_meteringsenders_crd.yaml
#operator.ibm.com_meterings_crd.yaml
#operator.ibm.com_meteringuis_crd.yaml

cp -p deploy/olm-catalog/ibm-metering-operator/ibm-metering-operator.package.yaml bundle/
cp -p deploy/olm-catalog/ibm-metering-operator/3.5.0/*.yaml bundle/

cd bundle
zip ibm-metering-metadata *.yaml
cd ..
