/*
Copyright 2022 KubeWorkz Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	resourcemanage "github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/handle"
	"github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/resources"
	"github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/resources/enum"
	"github.com/saashqdev/kubeworkz/pkg/clients"
	"github.com/saashqdev/kubeworkz/pkg/clog"
	"github.com/saashqdev/kubeworkz/pkg/utils/errcode"
)

func init() {
	resourcemanage.SetExtendHandler(enum.ExternalAccessAddressResourceType, addressHandle)
}

func addressHandle(param resourcemanage.ExtendContext) (interface{}, *errcode.ErrorInfo) {
	access := resources.NewSimpleAccess(param.Cluster, param.Username, param.Namespace)
	if allow := access.AccessAllow("", "services", "list"); !allow {
		return nil, errcode.ForbiddenErr
	}
	kubernetes := clients.Interface().Kubernetes(param.Cluster)
	if kubernetes == nil {
		return nil, errcode.ClusterNotFoundError(param.Cluster)
	}
	externalAccess := NewExternalAccess(kubernetes.Direct(), param.Namespace, param.ResourceName, param.FilterCondition, param.NginxNamespace, param.NginxTcpServiceConfigMap, param.NginxUdpServiceConfigMap)
	return externalAccess.getExternalIP()
}

// getExternalIP ingress-nginx pod status host IPs
func (s *ExternalAccess) getExternalIP() ([]string, *errcode.ErrorInfo) {
	var podList v1.PodList
	nginxLabel := map[string]string{
		"app.kubernetes.io/component": "controller",
		"app.kubernetes.io/name":      "ingress-nginx",
	}

	err := s.client.List(s.ctx, &podList, &client.ListOptions{Namespace: s.NginxNamespace, LabelSelector: labels.SelectorFromSet(nginxLabel)})
	if err != nil {
		clog.Error("can not find pod ingress-nginx in %s from cluster, %v", s.NginxNamespace, err)
		return nil, errcode.BadRequest(err)
	}

	var hostIps []string
	for _, pod := range podList.Items {
		if pod.Status.HostIP != "" {
			hostIps = append(hostIps, pod.Status.HostIP)
		}
	}

	return hostIps, nil
}