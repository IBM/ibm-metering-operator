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

# Get the SHA from the RepoDigests section of an operand image.
# Do "docker login" before running this script.
# Run this script from the parent dir by typing "scripts/get-image-sha.sh"


FILE=deploy/operator.yaml

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
   echo "   DM quay.io/opencloudio/metering-data-manager 3.5.0"
   echo "   UI quay.io/opencloudio/metering-ui 3.5.0"
   echo "   MCMUI quay.io/opencloudio/metering-mcmui 3.5.0"
   echo "   REPORT quay.io/opencloudio/metering-report 3.5.0"
   exit 1
fi

# check the TYPE value
if ! [[ ${TYPE_LIST[$TYPE]} ]]
then
   echo "$TYPE is not valid. Must be DM, UI, MCMUI, or REPORT"
   exit 1
fi

# pull the image
IMAGE="$NAME:$TAG"
echo "Pulling image $IMAGE"
docker pull $IMAGE

# get the SHA for the image
DIGEST="$(docker image inspect --format='{{index .RepoDigests 0}}' $IMAGE )"

# DIGEST should look like this: quay.io/opencloudio/metering-data-manager@sha256:nnnnnnnnnnnn
IFS='@'; ARRAY=($DIGEST); unset IFS
if [[ ${#ARRAY[@]} < 2 ]]
then
   echo "Cannot find SHA (@sha256:nnnnnnnnnnnn) in digest: $DIGEST"
   exit 1
fi
SHA=${ARRAY[1]}
echo "SHA=$SHA"

# delete the "name" and "value" lines for the old SHA
# for example:
#     - name: IMAGE_SHA_FOR_DM
#       value: "@sha256:10a844ffaf7733176e927e6c4faa04c2bc4410cf4d4ef61b9ae5240aa62d1456"

sed -i "/name: IMAGE_SHA_FOR_$TYPE/{N;d;}" $FILE

# insert the new SHA lines
LINE1="\            - name: IMAGE_SHA_FOR_$TYPE"
LINE2="\              value: \"@$SHA\""
sed -i "/DO NOT DELETE. Add image SHAs here/a $LINE1\n$LINE2" $FILE
