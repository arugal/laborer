#!/usr/bin/env bash

#
# Copyright 2020 arugal.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
set -ex

setup::cert_manager() {
    kubectl create namespace cert-manager
    helm repo add jetstack https://charts.jetstack.io
    helm repo update
    helm install cert-manager jetstack/cert-manager -n cert-manager --version v1.1.0 --wait --set installCRDs=true
}

setup::laborer() {
  make docker-build
  kind load docker-image controller:latest

  make manifests
  kustomize build config/test | kubectl apply --wait=${DEPLOY_WAIT} -f -

  kubectl get all -n laborer-system
  kubectl wait --for=condition=available deployment.apps/$(kubectl get deployment -n laborer-system | awk '$1 !~ /NAME/ {print $1}') -n laborer-system --timeout=1200s
  kubectl logs $(kubectl get pods -n laborer-system | awk '$1 !~ /NAME/ {print $1}') -n laborer-system
}

setup::test_image() {
  docker build -f test/docker/nginx1/Dockerfile -t web:v1 .
  kind load docker-image web:v1

  docker build -f test/docker/nginx2/Dockerfile -t web:v2 .
  kind load docker-image web:v2
}

setup::test_tool() {
  make test-tool-build

  sudo mv ./bin/test-tool /usr/local/bin/test-tool
}

case "$1" in
  cert_manager)
    setup::cert_manager
    ;;
  laborer)
    setup::laborer
    ;;
  test_image)
    setup::test_image
    ;;
  test_tool)
    setup::test_tool
    ;;
esac


