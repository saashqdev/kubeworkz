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

package register

import (
	_ "github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/resources/configmap"
	_ "github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/resources/cronjob"
	_ "github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/resources/deployment"
	_ "github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/resources/job"
	_ "github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/resources/node"
	_ "github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/resources/pod"
	_ "github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/resources/pvc"
	_ "github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/resources/replicaset"
	_ "github.com/saashqdev/kubeworkz/pkg/apiserver/kubeapi/resourcemanage/resources/service"
)
