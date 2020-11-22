#!/usr/bin/env bash

# ----------------------------------------------------------------------------
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.
# ----------------------------------------------------------------------------
# Copy by https://github.com/apache/skywalking/blob/master/test/e2e-mesh/e2e-istio/scripts/pre.sh
set -ex

K8S_VERSION=${K8S_VERSION:-'k8s-v1.19.1'}
KIND_VERSION=${KIND_VERSION:-'v0.9.0'}
KUSTOMIZE_VERSION=${KUSTOMIZE_VERSION:-'v3.8.7'}

# kubectl
curl -sSL "https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION#k8s-}/bin/linux/amd64/kubectl" -o /tmp/kubectl
chmod +x /tmp/kubectl
sudo mv /tmp/kubectl /usr/local/bin/kubectl

# kind
curl -sSL "https://kind.sigs.k8s.io/dl/${KIND_VERSION}/kind-linux-amd64" -Lo /tmp/kind
chmod +x /tmp/kind
sudo mv /tmp/kind /usr/local/bin/kind

# kustomize
curl -sSL "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2F${KUSTOMIZE_VERSION}/kustomize_${KUSTOMIZE_VERSION}_linux_amd64.tar.gz" -o /tmp/kustomize_linx.tar.gz
pushd /tmp
tar -zxvf ./kustomize_linx.tar.gz && chmod +x ./kustomize
sudo mv ./kustomize /usr/local/bin/kustomize
popd
