#!/usr/bin/env bash

#Copyright 2024 KubeWorkz Authors
#
#Licensed under the Apache License, Version 2.0 (the "License");
#you may not use this file except in compliance with the License.
#You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
#Unless required by applicable law or agreed to in writing, software
#distributed under the License is distributed on an "AS IS" BASIS,
#WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#See the License for the specific language governing permissions and
#limitations under the License.

function warden_webhook() {
cat >/etc/kube/warden/webhook.config <<EOF
apiVersion: v1
kind: Config
clusters:
  - name: warden
    cluster:
      server: https://127.0.0.1:31443/api/v1/warden/authenticate
      insecure-skip-tls-verify: true
users:
  - name: api-server

current-context: webhook
contexts:
  - context:
      cluster: warden
      user: api-server
    name: webhook
EOF
}

function audit_webhook() {
cat >/etc/kube/audit/audit-webhook.config  <<EOF
apiVersion: v1
clusters:
- cluster:
    server: http://127.0.0.1:30008/api/v1/kube/audit/k8s
    insecure-skip-tls-verify: true
  name: audit
contexts:
- context:
    cluster: audit
    user: ""
  name: default-context
current-context: default-context
kind: Config
preferences: {}
users: []
EOF
}

function audit_policy() {
cat >/etc/kube/audit/audit-policy.yaml  <<EOF
apiVersion: audit.k8s.io/v1
kind: Policy
omitStages:
  - "ResponseStarted"
  - "RequestReceived"
rules:
- level: None
  nonResourceURLs:
    - /apis*
    - /api/v1?timeout=*
    - /api?timeout=*
- level: Metadata
  userGroups: ["kubeworkz"]
EOF
}

function kind_config() {
cat >/etc/kube/kind/config.yaml <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraMounts:
  - hostPath: /etc/kube/
    containerPath: /etc/kube
  - hostPath: /var/log/
    containerPath: /var/log
  kubeadmConfigPatches:
  - |
    kind: ClusterConfiguration
    apiServer:
        extraArgs:
          authentication-token-webhook-config-file: "/etc/kube/warden/webhook.config"
          audit-policy-file: "/etc/kube/audit/audit-policy.yaml"
          audit-webhook-config-file: "/etc/kube/audit/audit-webhook.config"
          audit-log-path: "/var/log/audit"
          audit-log-maxage: "10"
          audit-log-maxsize: "100"
          audit-log-maxbackup: "10"
          audit-log-format: "json"
        extraVolumes:
        - name: "kube"
          hostPath: "/etc/kube"
          mountPath: "/etc/kube"
          readOnly: true
          pathType: DirectoryOrCreate
        - name: audit-log
          hostPath: "/var/log/audit"
          mountPath: "/var/log/audit"
          readOnly: false
          pathType: DirectoryOrCreate
EOF
}

mkdir -p /etc/kube/warden
mkdir -p /etc/kube/audit
mkdir -p /etc/kube/kind

warden_webhook
audit_webhook
audit_policy
kind_config

echo -e "\033[32m================================================\033[0m"
echo -e "\033[32m make configurations success\033[0m"