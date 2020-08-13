#!/bin/bash
############################################################################
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
############################################################################

# Get the SHA from an operand image or an operator image and put it in operator.yaml and the CSV file.
# Do "docker login" before running this script.
# Run this script from the parent dir by typing "scripts/get-image-sha.sh"

SED="sed"
unamestr=$(uname)
if [[ "$unamestr" == "Darwin" ]] ; then
    SED=gsed
    type $SED >/dev/null 2>&1 || {
        echo >&2 "$SED it's not installed. Try: brew install gnu-sed" ;
        exit 1;
    }
fi

# operator info
OPER_NAME=ibm-metering-operator
QUAY_REPO=quay.io/opencloudio

# check the input parms
TYPE=$1
NAME=$2
TAG=$3
if [[ $TAG == "" ]]; then
   echo "Missing parm. Need image type, image name and image tag"
   echo "Examples:"
   echo "   DM hyc-cloud-private-daily-docker-local.artifactory.swg-devops.com/ibmcom/metering-data-manager 3.6.0"
   echo "   UI hyc-cloud-private-daily-docker-local.artifactory.swg-devops.com/ibmcom/metering-ui 3.6.0"
   echo "   MCMUI hyc-cloud-private-daily-docker-local.artifactory.swg-devops.com/ibmcom/metering-mcmui 3.6.0"
   echo "   REPORT hyc-cloud-private-daily-docker-local.artifactory.swg-devops.com/ibmcom/metering-report 3.6.0"
   echo "   OPER hyc-cloud-private-daily-docker-local.artifactory.swg-devops.com/ibmcom/ibm-metering-operator 3.7.0"
   exit 1
fi

# check the TYPE value
if [[ $TYPE != "DM" && $TYPE != "UI" && $TYPE != "MCMUI" && $TYPE != "REPORT" && $TYPE != "OPER" ]]; then
   echo "$TYPE is not valid. Must be DM, UI, MCMUI, REPORT, or OPER"
   exit 1
fi

# check the TYPE value (doesn't work on mac)

# array of valid TYPE values
#declare -A TYPE_LIST
#TYPE_LIST[DM]=1
#TYPE_LIST[UI]=1
#TYPE_LIST[MCMUI]=1
#TYPE_LIST[REPORT]=1
#if ! [[ ${TYPE_LIST[$TYPE]} ]]
#then
#   echo "$TYPE is not valid. Must be DM, UI, MCMUI, or REPORT"
#   exit 1
#fi

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

if [[ $TYPE == "OPER" ]]; then
   echo "Updating operator image SHA in $OPER_FILE"
   # if the image is in "tag" format (name:tag) change it to "SHA" format (name@sha256:nnnn) first.
   # there are 10 blanks in front of "image:"
   $SED -e "s|          image: \(.*\)/${OPER_NAME}:\(.*\)|          image: \1/${OPER_NAME}@sha256:nnnn|" -i "${OPER_FILE}"

   # since the install process uses an ImageContentSourcePolicy to mirror quay.io/opencloudio to 
   # hyc-cloud-private-daily-docker-local.artifactory.swg-devops.com/ibmcom, we get the SHA from
   # a "daily" image but use QUAY_REPO as the image repo in operator.yaml.
   # there are 10 blanks in front of "image:"
   $SED -e "s|          image: \(.*\)/${OPER_NAME}@\(.*\)|          image: ${QUAY_REPO}/${OPER_NAME}@${SHA}|" -i "${OPER_FILE}"
else
   # delete the "name" and "value" lines for the old SHA
   # for example:
   #     - name: IMAGE_SHA_OR_TAG_DM
   #       value: sha256:10a844ffaf7733176e927e6c4faa04c2bc4410cf4d4ef61b9ae5240aa62d1456
   echo "Updating operand IMAGE_SHA_OR_TAG_$TYPE in $OPER_FILE"
   $SED -i "/name: IMAGE_SHA_OR_TAG_$TYPE/{N;d;}" "$OPER_FILE"

   # insert the new SHA lines
   LINE1="\            - name: IMAGE_SHA_OR_TAG_$TYPE"
   LINE2="\              value: $SHA"
   $SED -i "/env:/a $LINE1\n$LINE2" "$OPER_FILE"
fi

#---------------------------------------------------------
# update the CSV
#---------------------------------------------------------
CSV_FILE=deploy/olm-catalog/ibm-metering-operator/${CSV_VER}/ibm-metering-operator.v${CSV_VER}.clusterserviceversion.yaml

if [[ $TYPE == "OPER" ]]; then
   echo "Updating operator image SHA in $CSV_FILE"
   # if the image is in "tag" format (name:tag) change it to "SHA" format (name@sha256:nnnn) first.
   # there are 16 blanks in front of "image:"
   $SED -e "s|                image: \(.*\)/${OPER_NAME}:\(.*\)|                image: \1/${OPER_NAME}@sha256:nnnn|" -i "${CSV_FILE}"

   # since the install process uses an ImageContentSourcePolicy to mirror quay.io/opencloudio to 
   # hyc-cloud-private-daily-docker-local.artifactory.swg-devops.com/ibmcom, we get the SHA from
   # a "daily" image but use QUAY_REPO as the image repo in operator.yaml.
   # there are 16 blanks in front of "image:"
   $SED -e "s|                image: \(.*\)/${OPER_NAME}@\(.*\)|                image: ${QUAY_REPO}/${OPER_NAME}@${SHA}|" -i "${CSV_FILE}"
else
   # delete the "name" and "value" lines for the old SHA
   # for example:
   #     - name: IMAGE_SHA_OR_TAG_DM
   #       value: sha256:10a844ffaf7733176e927e6c4faa04c2bc4410cf4d4ef61b9ae5240aa62d1456
   echo "Updating operand IMAGE_SHA_OR_TAG_$TYPE in $CSV_FILE"
   $SED -i "/name: IMAGE_SHA_OR_TAG_$TYPE/{N;d;}" "$CSV_FILE"

   # insert the new SHA lines. need 4 more leading spaces compared to operator.yaml
   LINE1="\                - name: IMAGE_SHA_OR_TAG_$TYPE"
   LINE2="\                  value: $SHA"
   $SED -i "/env:/a $LINE1\n$LINE2" "$CSV_FILE"
fi
