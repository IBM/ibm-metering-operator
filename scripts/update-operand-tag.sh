#!/bin/bash
#
# Copyright 2020 IBM Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Update the required files with the new image tag for an operand.
#
# Run this script from the parent dir by typing "scripts/update-operand-tags.sh".
# Run this script for each of the operands.
#

SED="sed"
unamestr=$(uname)
if [[ "$unamestr" == "Darwin" ]] ; then
    SED=gsed
    type $SED >/dev/null 2>&1 || {
        echo >&2 "$SED it's not installed. Try: brew install gnu-sed" ;
        exit 1;
    }
fi

# check the input parms
if [ -z "${3}" ]
then
   echo "Missing parm. Need CSV version, image type, and image tag"
   echo "Usage:   $0 [CSV version] [image type] [image tag]"
   echo "         $0 3.6.3 DM 3.5.1"
   echo "         $0 3.6.3 UI 3.5.1"
   echo "         $0 3.6.3 MCMUI 3.5.2"
   echo "         $0 3.6.3 REPORT 3.5.1"
   exit 1
fi

OPERATOR_NAME=ibm-metering-operator
CSV_VERSION=${1}
TYPE=${2}
IMAGE_TAG=${3}

# check the TYPE value
if [[ $TYPE != "DM" && $TYPE != "UI" && $TYPE != "MCMUI" && $TYPE != "REPORT" ]]; then
   echo "$TYPE is not valid. Must be DM, UI, MCMUI, or REPORT"
   exit 1
fi


#---------------------------------------------------------
# verify that the CSV file exists
#---------------------------------------------------------
CSV_DIR=deploy/olm-catalog/${OPERATOR_NAME}/${CSV_VERSION}
CSV_FILE=${CSV_DIR}/${OPERATOR_NAME}.v${CSV_VERSION}.clusterserviceversion.yaml

if ! [ -f "${CSV_FILE}" ]; then
    echo "[WARN] ${CSV_FILE} does not exist. Check the version input parm."
    exit 1
fi

echo "******************************************"
echo " CSV_VERSION:  $CSV_VERSION"
echo " CSV_FILE:     $CSV_FILE"
echo " TYPE:         $TYPE"
echo " IMAGE_TAG:    $IMAGE_TAG"
echo "******************************************"
echo "Does the above look correct? (y/n) "
read -r ANSWER
if [[ "$ANSWER" != "y" ]]
then
  echo "Not going to update operand versions"
  exit 1
fi

#---------------------------------------------------------
# update operator.yaml and CSV file
#---------------------------------------------------------
OPERATOR_YAML=deploy/operator.yaml
if ! [ -f "${OPERATOR_YAML}" ]; then
    echo "[WARN] ${OPERATOR_YAML} does not exist."
    exit 1
fi

# delete the "name" and "value" lines for the old tag
# for example:
#     - name: IMAGE_SHA_OR_TAG_DM
#       value: 3.5.1
$SED -i "/name: IMAGE_SHA_OR_TAG_$TYPE/{N;d;}" "$OPERATOR_YAML"
$SED -i "/name: IMAGE_SHA_OR_TAG_$TYPE/{N;d;}" "$CSV_FILE"

# insert the new tag lines in operator.yaml 
echo -e "\n[INFO] Updating image tag for $TYPE in operator.yaml"
LINE1="\            - name: IMAGE_SHA_OR_TAG_$TYPE"
LINE2="\              value: $IMAGE_TAG"
$SED -i "/env:/a $LINE1\n$LINE2" "$OPERATOR_YAML"

# insert the new tag lines in the CSV file
# need 4 more leading spaces compared to operator.yaml
echo -e "\n[INFO] Updating image tag for $TYPE in CSV file"
LINE3="\                - name: IMAGE_SHA_OR_TAG_$TYPE"
LINE4="\                  value: $IMAGE_TAG"
$SED -i "/env:/a $LINE3\n$LINE4" "$CSV_FILE"

#---------------------------------------------------------
# update containers.go
#---------------------------------------------------------
CONTAINER_GO="pkg/resources/containers.go"
if ! [ -f "${CONTAINER_GO}" ]; then
    echo "[WARN] ${CONTAINER_GO} does not exist."
    exit 1
fi

if [[ $TYPE == "DM" ]]; then TYPE2="Dm"; fi
if [[ $TYPE == "UI" ]]; then TYPE2="UI"; fi
if [[ $TYPE == "MCMUI" ]]; then TYPE2="McmUI"; fi
if [[ $TYPE == "REPORT" ]]; then TYPE2="Report"; fi

echo -e "\n[INFO] Updating image tag for $TYPE in ${CONTAINER_GO}"
CONSTANT_NAME="Default${TYPE2}ImageTag"
$SED -i "s|const ${CONSTANT_NAME} = \"\([0-9]\.[0-9]\.[0-9]\)\"|const ${CONSTANT_NAME} = \"${IMAGE_TAG}\"|" "${CONTAINER_GO}"

#---------------------------------------------------------
# update Makefile
#---------------------------------------------------------
MAKEFILE="Makefile"
if ! [ -f "${MAKEFILE}" ]; then
    echo "[WARN] ${MAKEFILE} does not exist."
    exit 1
fi

echo -e "\n[INFO] Updating env var for $TYPE in ${MAKEFILE}"
VAR_NAME="OPERAND_TAG_${TYPE}"
$SED -i "s|${VAR_NAME} ?= \([0-9]\.[0-9]\.[0-9]\)|${VAR_NAME} ?= ${IMAGE_TAG}|" "${MAKEFILE}"
