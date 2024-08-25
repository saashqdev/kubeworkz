/*
Copyright 2024 Kubeworkz Authors

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

package pvc

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	resourcemanage "github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/handle"
	"github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/resources"
	"github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/resources/enum"
	"github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/resources/pod"
	"github.com/saashqdev/kubeworkz/pkg/clients"
	mgrclient "github.com/saashqdev/kubeworkz/pkg/multicluster/client"
	"github.com/saashqdev/kubeworkz/pkg/utils/errcode"
	"github.com/saashqdev/kubeworkz/pkg/utils/filter"
)

type Pvc struct {
	ctx             context.Context
	client          mgrclient.Client
	namespace       string
	filterCondition *filter.Condition
}

func init() {
	resourcemanage.SetExtendHandler(enum.PvcWorkLoadResourceType, handle)
}

func handle(param resourcemanage.ExtendContext) (interface{}, *errcode.ErrorInfo) {
	access := resources.NewSimpleAccess(param.Cluster, param.Username, param.Namespace)
	if allow := access.AccessAllow("", "persistentvolumeclaims", "list"); !allow {
		return nil, errcode.ForbiddenErr
	}
	kubernetes := clients.Interface().Kubernetes(param.Cluster)
	if kubernetes == nil {
		return nil, errcode.ClusterNotFoundError(param.Cluster)
	}
	pvc := NewPvc(kubernetes, param.Namespace, param.FilterCondition)
	return pvc.getPvcWorkloads(param.ResourceName)
}

func NewPvc(client mgrclient.Client, namespace string, condition *filter.Condition) Pvc {
	ctx := context.Background()
	return Pvc{
		ctx:             ctx,
		client:          client,
		namespace:       namespace,
		filterCondition: condition,
	}
}

// getPvcWorkloads get extend deployments
func (p *Pvc) getPvcWorkloads(pvcName string) (*unstructured.Unstructured, *errcode.ErrorInfo) {
	result := make(map[string]interface{})
	var pods []pod.ExtendPod
	var podList corev1.PodList
	err := p.client.Cache().List(p.ctx, &podList, client.InNamespace(p.namespace))
	if err != nil {
		return nil, errcode.BadRequest(err)
	}
	for _, item := range podList.Items {
		for _, volume := range item.Spec.Volumes {
			if volume.PersistentVolumeClaim == nil {
				continue
			}
			claimName := volume.PersistentVolumeClaim.ClaimName
			if claimName == pvcName {
				pods = append(pods, pod.ExtendPod{
					Reason: pod.GetPodReason(item),
					Pod:    item,
				})
				break
			}
		}
	}
	result["pods"] = pods
	result["total"] = len(pods)
	return &unstructured.Unstructured{Object: result}, nil
}
