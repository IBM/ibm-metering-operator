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

# Get the SHA from an operand image and put it in operator.yaml and the CSV file.
# Do "docker login" before running this script.
# Run this script from the parent dir by typing "scripts/get-image-sha.sh"


# array of valid TYPE values
declare -A TYPE_LIST
TYPE_LIST[DM]=1
TYPE_LIST[UI]=1
TYPE_LIST[MCMUI]=1
TYPE_LIST[REPORT]=1

# check the input parms
TYPE=$1
NAME=$2
TAG=$3
if [[ $TAG == "" ]]
then
   echo "Missing parm. Need image type, image name and image tag"
   echo "Examples:"
   echo "   DM quay.io/opencloudio/metering-data-manager 3.5.1"
   echo "   UI quay.io/opencloudio/metering-ui 3.5.1"
   echo "   MCMUI quay.io/opencloudio/metering-mcmui 3.5.0"
   echo "   REPORT quay.io/opencloudio/metering-report 3.5.1"
   exit 1
fi

# check the TYPE value
if ! [[ ${TYPE_LIST[$TYPE]} ]]
then
   echo "$TYPE is not valid. Must be DM, UI, MCMUI, or REPORT"
   exit 1
fi

#---------------------------------------------------------
# pull the image
#---------------------------------------------------------
IMAGE="$NAME:$TAG"
echo "Pulling image $IMAGE"
docker pull "$IMAGE"

#---------------------------------------------------------
# get the SHA for the image
#---------------------------------------------------------
DIGEST="$(docker images --digests "$NAME" | grep "$TAG" | awk 'FNR==1{print $3}')"

# DIGEST should look like this: sha256:10a844ffaf7733176e927e6c4faa04c2bc4410cf4d4ef61b9ae5240aa62d1456
if [[ $DIGEST != sha256* ]]
then
    echo "Cannot find SHA (sha256:nnnnnnnnnnnn) in digest: $DIGEST"
    exit 1
fi

SHA=$DIGEST
echo "SHA=$SHA"

#---------------------------------------------------------
# get the operator version number
#---------------------------------------------------------
VER_LINE="$(grep 'Version =' version/version.go )"

# VER_LINE should look like this: Version = "3.6.1"
# split the line on the double quote
IFS='"'; read -r -a ARRAY <<< "$VER_LINE"; unset IFS
if [[ ${#ARRAY[@]} -lt 2 ]]
then
   echo "Cannot find version number in version/version.go"
   exit 1
fi
# the version number is the second element in the array
CSV_VER=${ARRAY[1]}
echo "CSV_VER=$CSV_VER"

#---------------------------------------------------------
# update operator.yaml
#---------------------------------------------------------
OPER_FILE=deploy/operator.yaml

# delete the "name" and "value" lines for the old SHA
# for example:
#     - name: IMAGE_SHA_OR_TAG_DM
#       value: sha256:10a844ffaf7733176e927e6c4faa04c2bc4410cf4d4ef61b9ae5240aa62d1456

sed -i "/name: IMAGE_SHA_OR_TAG_$TYPE/{N;d;}" $OPER_FILE

# insert the new SHA lines
LINE1="\            - name: IMAGE_SHA_OR_TAG_$TYPE"
LINE2="\              value: $SHA"
sed -i "/env:/a $LINE1\n$LINE2" $OPER_FILE

#---------------------------------------------------------
# update the CSV
#---------------------------------------------------------
CSV_FILE=deploy/olm-catalog/ibm-metering-operator/${CSV_VER}/ibm-metering-operator.v${CSV_VER}.clusterserviceversion.yaml

# delete the "name" and "value" lines for the old SHA
# for example:
#     - name: IMAGE_SHA_OR_TAG_DM
#       value: sha256:10a844ffaf7733176e927e6c4faa04c2bc4410cf4d4ef61b9ae5240aa62d1456

sed -i "/name: IMAGE_SHA_OR_TAG_$TYPE/{N;d;}" $CSV_FILE

# insert the new SHA lines. need 4 more leading spaces compared to operator.yaml
LINE1="\                - name: IMAGE_SHA_OR_TAG_$TYPE"
LINE2="\                  value: $SHA"
sed -i "/env:/a $LINE1\n$LINE2" $CSV_FILE
