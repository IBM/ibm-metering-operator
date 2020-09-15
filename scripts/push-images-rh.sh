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
# Use this script to upload images for Red Hat certification
#
# ./push-images-rh.sh <image-name> <tag> <arch> <Registry Key from RH cert> <RH PID> <-scan>
#
# Go to the "images" URL (https://connect.redhat.com/project/4450911/images)
# - under Container Images -> PID, copy the ospid value for <RH PID>
# - select "Push Image Manually", goto View Registry Key and copy the Registry Key for <Registry Key from RH cert>
# OLD: go to the "view" URL (https://connect.redhat.com/project/4450911/images)
# - goto Upload Your Image -> View Registry Key and copy the Registry Key
# - goto Upload Your Image -> Tag Your Container and copy the ospid value from the "docker tag" cmd
#
# registry-key: <a very very long string, over 1000 chars>
# tag-cmd:      docker tag [image-id] scan.connect.redhat.com/<ospid value>/[image-name]:[tag]
#
# Use the "-scan" parm the first time to upload the image with a temp tag so you can verify the image will pass the scan.
# To make the command line parameters easier to manage, put the registry-key in an env var like REG_KEY
#   ./push-images-rh.sh <image-name> <tag> <arch> $REG_KEY <RH PID> <-scan>
#

IMAGE_NAME=$1
TAG=$2
ARCH=$3
PASSWORD=$4
RH_PID=$5
TEST=$6


# OLD: QUAY=quay.io/opencloudio/$IMAGE:$TAG
# LOCAL_IMAGE example: hyc-cloud-private-integration-docker-local.artifactory.swg-devops.com/ibmcom/metering-data-manager-ppc64le:3.6.0
# SCAN_IMAGE example: 
LOCAL_IMAGE=hyc-cloud-private-integration-docker-local.artifactory.swg-devops.com/ibmcom/$IMAGE_NAME-$ARCH:$TAG
SCAN_IMAGE=scan.connect.redhat.com/$RH_PID/$IMAGE_NAME:$TAG-$ARCH$TEST

echo "### This is going to pull your image from the repo"
echo docker pull "$LOCAL_IMAGE"
echo
echo "### This will log you into your RH project for THIS image"
echo docker login -u unused -p "$PASSWORD" scan.connect.redhat.com
echo
echo "### This will tag the image for RH scan"
echo docker tag "$LOCAL_IMAGE" "$SCAN_IMAGE"
echo
echo "### This will push to Red Hat... if this is the first scan there MUST be something appended to the end."
echo docker push "$SCAN_IMAGE"

echo
echo
echo "Does the above look correct? (y/n) "
read -r ANSWER
if [[ "$ANSWER" != "y" ]]
then
  echo "Not going to run commands"
  exit 1
fi

docker pull "$LOCAL_IMAGE"
docker login -u unused -p "$PASSWORD" scan.connect.redhat.com
docker tag "$LOCAL_IMAGE" "$SCAN_IMAGE"
docker push "$SCAN_IMAGE"
