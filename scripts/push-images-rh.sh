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
# ./push-images-rh.sh [image-name]:[tag] <Registry Key from RH cert> <Tag sample cmd from RH cert> <-scan>
#
# goto Upload Your Image -> View Registry Key and copy the Registry Key
# goto Upload Your Image -> Tag Your Container and copy the "docker tag" cmd
#
# registry-key: <a very very long string, over 1000 chars>
# tag-cmd:      docker tag [image-id] scan.connect.redhat.com/ospid-b05fd075-5d9a-421f-bab6-033a443b2d35/[image-name]:[tag]
#
# use the "-scan" parm the first time to upload the image with a temp tag so you can verify the image will pass the scan
#

IMAGE=$(echo "$1" | cut -d ':' -f1)
TAG=$(echo "$1" | cut -d ':' -f2)
PASSWORD=$2
REPO=$(echo "$6" | cut -d '/' -f2)
TEST=$7

QUAY=quay.io/opencloudio/$IMAGE:$TAG
REDHAT=scan.connect.redhat.com/$REPO/$IMAGE:$TAG$TEST

echo "### This is going to pull your image from quay"
echo docker pull "$QUAY"
echo
echo "### This will log you into your RH project for THIS image"
echo docker login -u unused -p "$PASSWORD" scan.connect.redhat.com
echo
echo "### This will tag the quay image for redhat scan"
echo docker tag "$QUAY" "$REDHAT"
echo
echo "### This will push to redhat...if this is the first scan there MUST be something appended to the end."
echo docker push "$REDHAT"

echo
echo
echo "Does the above look correct? (y/n) "
read -r ANSWER
if [[ "$ANSWER" != "y" ]]
then
  echo "Not going to run commands"
  exit 1
fi

docker pull "$QUAY"
docker login -u unused -p "$PASSWORD" scan.connect.redhat.com
docker tag "$QUAY" "$REDHAT"
docker push "$REDHAT"
