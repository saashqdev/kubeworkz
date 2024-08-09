/*
Copyright 2024 KubeWorkz Authors

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

package resourcemanage

import (
	"net/http"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"

	clusterv1 "github.com/saashqdev/kubeworkz/pkg/apis/cluster/v1"
	tenantv1 "github.com/saashqdev/kubeworkz/pkg/apis/tenant/v1"
	"github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/resources"
	"github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/resources/enum"
	"github.com/saashqdev/kubeworkz/pkg/apiserver/middlewares/audit"
	"github.com/saashqdev/kubeworkz/pkg/clients"
	"github.com/saashqdev/kubeworkz/pkg/clog"
	"github.com/saashqdev/kubeworkz/pkg/utils/constants"
	"github.com/saashqdev/kubeworkz/pkg/utils/env"
	"github.com/saashqdev/kubeworkz/pkg/utils/errcode"
	"github.com/saashqdev/kubeworkz/pkg/utils/response"
)

var (
	extendFuncMap = make(map[string]ExtendFunc)
)

type ExtendHandler struct {
	auditHandler             *audit.Handler
	NginxNamespace           string
	NginxTcpServiceConfigMap string
	NginxUdpServiceConfigMap string
}

func NewExtendHandler(namespace string, tcpCm string, udpCm string) *ExtendHandler {
	return &ExtendHandler{
		NginxNamespace:           namespace,
		NginxTcpServiceConfigMap: tcpCm,
		NginxUdpServiceConfigMap: udpCm,
	}
}

type ExtendFunc func(extendCtx ExtendContext) (interface{}, *errcode.ErrorInfo)

// SetExtendHandler the func to register real handler func
func SetExtendHandler(resource enum.ResourceTypeEnum, extendFunc ExtendFunc) {
	extendFuncMap[string(resource)] = extendFunc
}

// ExtendHandle api/v1/kube/extend/clusters/{cluster}/namespaces/{namespace}/{resourceType}
func (e *ExtendHandler) ExtendHandle(c *gin.Context) {
	// request param
	cluster := c.Param("cluster")
	namespace := c.Param("namespace")
	resourceType := c.Param("resourceType")
	resourceName := c.Param("resourceName")
	condition := parseQueryParams(c)
	httpMethod := c.Request.Method

	// k8s client
	client := clients.Interface().Kubernetes(cluster)
	if client == nil {
		response.FailReturn(c, errcode.ClusterNotFoundError(cluster))
		return
	}
	// get user info
	username := c.GetString(constants.UserName)

	extendCtx := ExtendContext{
		Cluster:                  cluster,
		Namespace:                namespace,
		ResourceName:             resourceName,
		FilterCondition:          condition,
		Action:                   httpMethod,
		Username:                 username,
		NginxNamespace:           e.NginxNamespace,
		NginxTcpServiceConfigMap: e.NginxTcpServiceConfigMap,
		NginxUdpServiceConfigMap: e.NginxUdpServiceConfigMap,
		AuditHandler:             e.auditHandler,
		GinContext:               c,
	}

	clog.Debug("request extend api with method (%v), cluster (%v), namespace (%v), resource type (%v), resource name (%v)", httpMethod, cluster, namespace, resourceType, resourceName)

	// get real handler func and work, if not found, return not support error
	if extendFunc, ok := extendFuncMap[resourceType]; ok {
		result, errInfo := extendFunc(extendCtx)
		if errInfo != nil {
			clog.Error("get extend res err, resourceType: %s, error: %s", resourceType, errInfo.Message)
			response.FailReturn(c, errInfo)
			return
		}
		response.SuccessReturn(c, result)
		return
	}
	response.FailReturn(c, errcode.InvalidResourceTypeErr)
	return
}

// GetPodContainerLog get k8s pod log and page log
// Deprecated: Use GetProxyPodContainerLog instead, get better performance
func GetPodContainerLog(c *gin.Context) {
	cluster := c.Param("cluster")
	namespace := c.Param("namespace")
	condition := parseQueryParams(c)
	// k8s client
	client := clients.Interface().Kubernetes(cluster)
	if client == nil {
		response.FailReturn(c, errcode.ClusterNotFoundError(cluster))
		return
	}
	// access
	username := c.GetString(constants.UserName)
	access := resources.NewSimpleAccess(cluster, username, namespace)
	if allow := access.AccessAllow("", "pods", "list"); !allow {
		response.FailReturn(c, errcode.ForbiddenErr)
		return
	}
	podLog := NewPodLog(client, namespace, condition)
	podLog.HandleLogs(c)
}

// GetProxyPodContainerLog get native k8s pod container log
func GetProxyPodContainerLog(c *gin.Context) {
	cluster := c.Param("cluster")
	namespace := c.Param("namespace")
	condition := parseQueryParams(c)
	// k8s client
	client := clients.Interface().Kubernetes(cluster)
	if client == nil {
		response.FailReturn(c, errcode.ClusterNotFoundError(cluster))
		return
	}
	// access
	username := c.GetString(constants.UserName)
	access := resources.NewSimpleAccess(cluster, username, namespace)
	if allow := access.AccessAllow("", "pods", "list"); !allow {
		response.FailReturn(c, errcode.ForbiddenErr)
		return
	}
	podLog := NewProxyPodLog(client, namespace, condition)
	podLog.HandleLogs(c)
}

// GetFeatureConfig shows layout of integrated components
// all users have read-only access ability
func GetFeatureConfig(c *gin.Context) {
	cli := clients.Interface().Kubernetes(constants.LocalCluster)
	if cli == nil {
		response.FailReturn(c, errcode.BadRequest(nil))
		return
	}

	cm := &corev1.ConfigMap{}
	key := types.NamespacedName{Name: "kubeworkz-feature-config", Namespace: env.CubeNamespace()}

	err := cli.Cache().Get(c.Request.Context(), key, cm)
	if err != nil {
		if errors.IsNotFound(err) {
			response.FailReturn(c, errcode.CustomReturn(http.StatusNotFound, "configmap(%v/%v) not found", key.Namespace, key.Name))
			return
		}
		response.FailReturn(c, errcode.BadRequest(err))
		return
	}

	response.SuccessReturn(c, cm.Data)
}

// GetConfigMap show system configMap
// all users have read-only access ability
func GetConfigMap(c *gin.Context) {
	cmName := c.Param("configmap")
	if cmName == "" {
		response.FailReturn(c, errcode.InvalidBodyFormat)
		return
	}

	cli := clients.Interface().Kubernetes(constants.LocalCluster)
	if cli == nil {
		response.FailReturn(c, errcode.BadRequest(nil))
		return
	}

	cm := &corev1.ConfigMap{}
	key := types.NamespacedName{Name: cmName, Namespace: env.CubeNamespace()}

	err := cli.Cache().Get(c.Request.Context(), key, cm)
	if err != nil {
		if errors.IsNotFound(err) {
			response.FailReturn(c, errcode.CustomReturn(http.StatusNotFound, "configmap(%v/%v) not found", key.Namespace, key.Name))
			return
		}
		response.FailReturn(c, errcode.BadRequest(err))
		return
	}

	response.SuccessReturn(c, cm.Data)
}

// IngressDomainSuffix Get Ingress Domain Suffix by cluster and project
func IngressDomainSuffix(c *gin.Context) {
	clusterName := c.Query("cluster")
	projectName := c.Query("project")
	client := clients.Interface().Kubernetes(constants.LocalCluster)
	if client == nil {
		clog.Error("get cluster failed")
		response.FailReturn(c, errcode.ClusterNotFoundError(clusterName))
		return
	}
	cluster := clusterv1.Cluster{}
	err := client.Cache().Get(c, types.NamespacedName{Name: clusterName}, &cluster)
	if err != nil {
		if errors.IsNotFound(err) {
			response.FailReturn(c, errcode.ClusterNotFoundError(clusterName))
			return
		}
		response.FailReturn(c, errcode.BadRequest(err))
		return
	}

	project := tenantv1.Project{}
	err = client.Cache().Get(c, types.NamespacedName{Name: projectName}, &project)
	if err != nil {
		if errors.IsNotFound(err) {
			response.FailReturn(c, errcode.CustomReturn(http.StatusNotFound, "project(%v) not found", projectName))
			return
		}
		response.FailReturn(c, errcode.BadRequest(err))
		return
	}

	// because the cluster ingress domain suffix may repeat to project ingress domain suffix,so we use set in here to deduplication
	tmpSet := sets.Set[string]{}
	if len(cluster.Spec.IngressDomainSuffix) != 0 {
		tmpSet.Insert(cluster.Spec.IngressDomainSuffix)
	}
	for _, suffix := range project.Spec.IngressDomainSuffix {
		tmpSet.Insert(suffix)
	}
	res := sets.List(tmpSet)

	response.SuccessReturn(c, res)
}