/*
Copyright 2022 Kubeworkz Authors

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

package node

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	resourcemanage "github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/handle"
	"github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/resources"
	"github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/resources/enum"
	"github.com/saashqdev/kubeworkz/pkg/clients"
	"github.com/saashqdev/kubeworkz/pkg/clog"
	mgrclient "github.com/saashqdev/kubeworkz/pkg/multicluster/client"
	"github.com/saashqdev/kubeworkz/pkg/utils/errcode"
	"github.com/saashqdev/kubeworkz/pkg/utils/filter"
)

type Node struct {
	ctx             context.Context
	client          mgrclient.Client
	filterCondition *filter.Condition
}

func init() {
	resourcemanage.SetExtendHandler(enum.NodeResourceType, handle)
}

func handle(param resourcemanage.ExtendContext) (interface{}, *errcode.ErrorInfo) {
	access := resources.NewSimpleAccess(param.Cluster, param.Username, param.Namespace)
	if allow := access.AccessAllow("", "nodes", "list"); !allow {
		return nil, errcode.ForbiddenErr
	}
	kubernetes := clients.Interface().Kubernetes(param.Cluster)
	if kubernetes == nil {
		return nil, errcode.ClusterNotFoundError(param.Cluster)
	}
	node := NewNode(kubernetes, param.FilterCondition)
	return node.GetExtendNodes()
}

func NewNode(client mgrclient.Client, condition *filter.Condition) Node {
	ctx := context.Background()
	return Node{
		ctx:             ctx,
		client:          client,
		filterCondition: condition,
	}
}

func (node *Node) GetExtendNodes() (*unstructured.Unstructured, *errcode.ErrorInfo) {
	resultMap := make(map[string]interface{})

	// get deployment list from k8s cluster
	nodeList := corev1.NodeList{}
	err := node.client.Cache().List(node.ctx, &nodeList)
	if err != nil {
		clog.Error("can not find node in cluster, %v", err)
		return nil, errcode.BadRequest(err)
	}
	// filterCondition list by selector/sort/page
	extendInfo := addExtendInfo(&nodeList)
	u := &unstructured.UnstructuredList{}
	u.Items = extendInfo
	total, err := filter.GetEmptyFilter().FilterObjectList(u, node.filterCondition)
	if err != nil {
		clog.Error("filterCondition nodeList error, err: %s", err.Error())
		return nil, errcode.BadRequest(err)
	}
	resultMap["total"] = total
	resultMap["items"] = u.Items
	return &unstructured.Unstructured{
		Object: resultMap,
	}, nil
}

func addExtendInfo(nodeList *corev1.NodeList) []unstructured.Unstructured {
	items := make([]unstructured.Unstructured, 0)
	for _, node := range nodeList.Items {
		// parse job status
		status := ParseNodeStatus(node)

		// add extend info
		extendInfo := make(map[string]interface{})
		extendInfo["status"] = status

		// add node info and extend info
		result := make(map[string]interface{})
		result["metadata"] = node.ObjectMeta
		result["spec"] = node.Spec
		result["status"] = node.Status
		result["extendInfo"] = extendInfo

		//add to list
		items = append(items, unstructured.Unstructured{Object: result})
	}
	return items
}

func ParseNodeStatus(node corev1.Node) (status string) {
	if node.Spec.Unschedulable {
		return unscheduledStatus
	}
	if node.Status.Conditions == nil {
		return ""
	}
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
			return normal
		}
	}
	return abNormal
}
