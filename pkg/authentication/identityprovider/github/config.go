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

package github

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/saashqdev/kubeworkz/pkg/authentication"
	"github.com/saashqdev/kubeworkz/pkg/clients"
	"github.com/saashqdev/kubeworkz/pkg/clog"
	"github.com/saashqdev/kubeworkz/pkg/utils/constants"
	"github.com/saashqdev/kubeworkz/pkg/utils/env"
	"github.com/saashqdev/kubeworkz/pkg/warden/localmgr/controllers/hotplug"
)

const configMapName = "kubeworkz-auth-config"

func getConfig() authentication.GitHubConfig {
	var gitHubConfig authentication.GitHubConfig

	kClient := clients.Interface().Kubernetes(constants.LocalCluster).Cache()
	if kClient == nil {
		clog.Error("get pivot cluster client is nil")
		return gitHubConfig
	}
	cm := &v1.ConfigMap{}
	err := kClient.Get(context.Background(), client.ObjectKey{Name: configMapName, Namespace: env.KubeNamespace()}, cm)
	if err != nil {
		clog.Error("get configmap from K8s err: %v", err)
		return gitHubConfig
	}

	config := cm.Data["github"]
	if config == "" {
		clog.Error("github config is nil")
		return gitHubConfig
	}
	configJson, err := hotplug.YamlStringToJson(config)
	if err != nil {
		clog.Error("%v", err.Error())
		return gitHubConfig
	}
	if configJson["enabled"] != nil && configJson["enabled"].(bool) == true {
		gitHubConfig.GitHubIsEnable = true
	} else {
		return gitHubConfig
	}

	if configJson["clientId"] != nil {
		gitHubConfig.ClientID = configJson["clientId"].(string)
	}
	if configJson["clientSecret"] != nil {
		gitHubConfig.ClientSecret = configJson["clientSecret"].(string)
	}
	return gitHubConfig
}
