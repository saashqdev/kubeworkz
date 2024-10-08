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

package webhooks

import (
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	clusterWebhook "github.com/saashqdev/kubeworkz/pkg/ctrlmgr/webhooks/cluster"
	hotplugWebhook "github.com/saashqdev/kubeworkz/pkg/ctrlmgr/webhooks/hotplug"
	"github.com/saashqdev/kubeworkz/pkg/ctrlmgr/webhooks/project"
	"github.com/saashqdev/kubeworkz/pkg/ctrlmgr/webhooks/quota"
	"github.com/saashqdev/kubeworkz/pkg/ctrlmgr/webhooks/tenant"
)

// todo: change set func if need

// SetupWithWebhooks set up webhooks into manager
func SetupWithWebhooks(mgr manager.Manager) {
	hookServer := mgr.GetWebhookServer()

	client := mgr.GetClient()
	decoder := admission.NewDecoder(mgr.GetScheme())
	hookServer.Register("/validate-cluster-kubeworkz-io-v1-cluster", admission.ValidatingWebhookFor(mgr.GetScheme(), clusterWebhook.NewClusterValidator(client)))
	hookServer.Register("/validate-hotplug-kubeworkz-io-v1-hotplug", admission.ValidatingWebhookFor(mgr.GetScheme(), hotplugWebhook.NewHotplugValidator(client)))
	hookServer.Register("/validate-quota-kubeworkz-io-v1-kube-resource-quota", &webhook.Admission{Handler: quota.NewKubeResourceQuotaValidator(client, decoder)})
	hookServer.Register("/validate-tenant-kubeworkz-io-v1-tenant", &webhook.Admission{Handler: tenant.NewValidator(decoder)})
	hookServer.Register("/validate-tenant-kubeworkz-io-v1-project", &webhook.Admission{Handler: project.NewValidator(client, decoder)})
}
