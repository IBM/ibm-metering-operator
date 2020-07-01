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
# registry-key: eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJvc3BpZC1iMDVmZDA3NS01ZDlhLTQyMWYtYmFiNi0wMzNhNDQzYjJkMzUiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlY3JldC5uYW1lIjoib3NwaWQtYjA1ZmQwNzUtNWQ5YS00MjFmLWJhYjYtMDMzYTQ0M2IyZDM1LXNhLXRva2VuLW44c3ZtIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQubmFtZSI6Im9zcGlkLWIwNWZkMDc1LTVkOWEtNDIxZi1iYWI2LTAzM2E0NDNiMmQzNS1zYSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6ImY3MTM0ZGZjLTY5NWEtMTFlYS05NzQ1LTBhODk4ODEyYzkzMSIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpvc3BpZC1iMDVmZDA3NS01ZDlhLTQyMWYtYmFiNi0wMzNhNDQzYjJkMzU6b3NwaWQtYjA1ZmQwNzUtNWQ5YS00MjFmLWJhYjYtMDMzYTQ0M2IyZDM1LXNhIn0.bQ2lZN4EUb9OHdDAJydoMeTFj-kCgGfBfO62L9SxKPhdkjuidt1IMuCKbcRVbQDYR9R3heMZnfJ1-GvKPrwXeo6SyBGAdVsV-pRS1cxhvEWQeXebfKhq8EEQ8cCk6yCs3mDznlQLse04iW00NeYwE8iBRJP6BlmH7Q3yrZU3b2xan4lwHNa30UThQS2ovztVDKNXps4PlvE--GXc5kfYRatA9D-hu4FIIklBRfOwTwu6BpQvkLyruG5jCANJKiWYW8yr7Fg3JVD6_8x68wZTdzziy1pLD-hg-bz5xQYj42r88wCwBWI31r6d8OlJQlnp-XJmLUFtpOx8MwaA_xh-FA
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
