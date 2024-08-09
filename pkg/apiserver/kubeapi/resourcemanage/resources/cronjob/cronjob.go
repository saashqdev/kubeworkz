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

package cronjob

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	resourcemanage "github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/handle"
	"github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/resources"
	"github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/resources/enum"
	jobRes "github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/resources/job"
	"github.com/saashqdev/kubeworkz/pkg/clients"
	"github.com/saashqdev/kubeworkz/pkg/clog"
	"github.com/saashqdev/kubeworkz/pkg/conversion"
	"github.com/saashqdev/kubeworkz/pkg/utils/errcode"
	"github.com/saashqdev/kubeworkz/pkg/utils/filter"
)

type CronJob struct {
	ctx             context.Context
	client          client.Client
	cache           cache.Cache
	namespace       string
	filterCondition *filter.Condition
}

func init() {
	resourcemanage.SetExtendHandler(enum.CronResourceType, handle)
}

func handle(param resourcemanage.ExtendContext) (interface{}, *errcode.ErrorInfo) {
	access := resources.NewSimpleAccess(param.Cluster, param.Username, param.Namespace)
	if allow := access.AccessAllow("batch", "cronjobs", "list"); !allow {
		return nil, errcode.ForbiddenErr
	}
	kubernetes := clients.Interface().Kubernetes(param.Cluster)
	if kubernetes == nil {
		return nil, errcode.ClusterNotFoundError(param.Cluster)
	}
	convertor, err := conversion.NewVersionConvertor(kubernetes.CacheDiscovery(), kubernetes.RESTMapper())
	if err != nil {
		return nil, errcode.BadRequest(err)
	}
	client := conversion.WrapClient(kubernetes.Direct(), convertor, true)
	cache := conversion.WrapCache(kubernetes.Cache(), convertor)
	cronjob := NewCronJob(client, cache, param.Namespace, param.FilterCondition)
	if param.ResourceName == "" {
		return cronjob.getExtendCronJobs()
	} else {
		return cronjob.getExtendCronJob(param.ResourceName)
	}
}

func NewCronJob(client client.Client, cache cache.Cache, namespace string, condition *filter.Condition) CronJob {
	ctx := context.Background()
	return CronJob{
		ctx:             ctx,
		client:          client,
		cache:           cache,
		namespace:       namespace,
		filterCondition: condition,
	}
}

// getExtendCronJobs get extend deployments
func (c *CronJob) getExtendCronJobs() (*unstructured.Unstructured, *errcode.ErrorInfo) {
	resultMap := make(map[string]interface{})

	// get deployment list from k8s cluster
	var cronJobList batchv1beta1.CronJobList
	err := c.cache.List(c.ctx, &cronJobList, client.InNamespace(c.namespace))
	if err != nil {
		clog.Error("can not find cronjob in %s from cluster, %v", c.namespace, err)
		return nil, errcode.BadRequest(err)
	}

	// filter list by selector/sort/page
	total, err := filter.GetEmptyFilter().FilterObjectList(&cronJobList, c.filterCondition)
	if err != nil {
		clog.Error("can not filter cronjob, err: %s", err.Error())
		return nil, errcode.BadRequest(err)
	}
	// add pod status info
	resultList := c.addExtendInfo(cronJobList)

	resultMap["total"] = total
	resultMap["items"] = resultList

	return &unstructured.Unstructured{Object: resultMap}, nil
}

// getExtendCronJob get extend deployments
func (c *CronJob) getExtendCronJob(name string) (*unstructured.Unstructured, *errcode.ErrorInfo) {
	// get deployment list from k8s cluster
	var cronJob batchv1beta1.CronJob
	err := c.cache.Get(c.ctx, types.NamespacedName{Namespace: c.namespace, Name: name}, &cronJob)
	if err != nil {
		clog.Error("can not find cronjob %s/%s from cluster, %v", c.namespace, name, err)
		return nil, errcode.BadRequest(err)
	}

	var cronJobList batchv1beta1.CronJobList
	cronJobList.Items = []batchv1beta1.CronJob{cronJob}
	resultList := c.addExtendInfo(cronJobList)
	if len(resultList) == 0 {
		return nil, errcode.BadRequest(fmt.Errorf("can not parse cronjob %s/%s", c.namespace, name))
	}

	return &resultList[0], nil
}

func (c *CronJob) addExtendInfo(cronJobList batchv1beta1.CronJobList) []unstructured.Unstructured {
	resultList := make([]unstructured.Unstructured, 0)
	jobArrMap := c.getOwnerJobs()
	for _, cronJob := range cronJobList.Items {
		// parse job status
		status := parseCronJobStatus(cronJob)
		jobArr, ok := jobArrMap[string(cronJob.UID)]
		runningJobCount := 0
		if ok {
			for _, job := range jobArr {
				extendInfo := job.(map[string]interface{})["extendInfo"]
				extendInfoStatus := extendInfo.(map[string]interface{})["status"].(string)
				if extendInfoStatus == "Running" {
					runningJobCount++
				}
			}
		}
		extendInfo := make(map[string]interface{})
		extendInfo["status"] = status
		extendInfo["runningJobCount"] = runningJobCount
		extendInfo["jobCount"] = len(jobArr)
		extendInfo["jobs"] = jobArr

		// create result map
		result := make(map[string]interface{})
		result["metadata"] = cronJob.ObjectMeta
		result["spec"] = cronJob.Spec
		result["status"] = cronJob.Status
		result["extendInfo"] = extendInfo
		res := unstructured.Unstructured{
			Object: result,
		}
		resultList = append(resultList, res)
	}

	return resultList
}

func (c *CronJob) getOwnerJobs() map[string][]interface{} {
	result := make(map[string][]interface{})
	var jobList batchv1.JobList
	err := c.cache.List(c.ctx, &jobList, client.InNamespace(c.namespace))
	if err != nil {
		clog.Error("can not find jobs from cluster, %v", err)
		return nil
	}

	for _, job := range jobList.Items {
		if len(job.OwnerReferences) == 0 {
			continue
		}
		uid := string(job.OwnerReferences[0].UID)

		status := jobRes.ParseJobStatus(job)
		extendInfo := make(map[string]interface{})
		extendInfo["status"] = status
		// create result map
		jobMap := make(map[string]interface{})
		jobMap["metadata"] = job.ObjectMeta
		jobMap["spec"] = job.Spec
		jobMap["status"] = job.Status
		jobMap["extendInfo"] = extendInfo

		if jobArr, ok := result[uid]; ok {
			jobArr = append(jobArr, jobMap)
			result[uid] = jobArr
		} else {
			var jobArrTemp []interface{}
			jobArrTemp = append(jobArrTemp, jobMap)
			result[uid] = jobArrTemp
		}
	}
	return result
}

func parseCronJobStatus(cronjob batchv1beta1.CronJob) (status string) {
	if *cronjob.Spec.Suspend {
		return "Fail"
	}
	return "Running"
}