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

package localmgr

import (
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	admisson "sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/saashqdev/kubeworkz/pkg/utils/ctrlopts"
	"github.com/saashqdev/kubeworkz/pkg/warden/localmgr/controllers/crds"
	"github.com/saashqdev/kubeworkz/pkg/warden/localmgr/controllers/hotplug"
	project "github.com/saashqdev/kubeworkz/pkg/warden/localmgr/controllers/project"
	"github.com/saashqdev/kubeworkz/pkg/warden/localmgr/controllers/quota"
	tenant "github.com/saashqdev/kubeworkz/pkg/warden/localmgr/controllers/tenant"
	user "github.com/saashqdev/kubeworkz/pkg/warden/localmgr/controllers/user"
	hotplug2 "github.com/saashqdev/kubeworkz/pkg/warden/localmgr/webhooks/hotplug"
	project2 "github.com/saashqdev/kubeworkz/pkg/warden/localmgr/webhooks/project"
	quota2 "github.com/saashqdev/kubeworkz/pkg/warden/localmgr/webhooks/quota"
	tenant2 "github.com/saashqdev/kubeworkz/pkg/warden/localmgr/webhooks/tenant"
)

// setupControllersWithManager set up controllers into manager
func setupControllersWithManager(m *LocalManager, controllers string) error {
	var err error

	ctrls := ctrlopts.ParseControllers(controllers)

	if ctrlopts.IsControllerEnabled("hotplug", ctrls) {
		err = hotplug.SetupWithManager(m.Manager, m.IsMemberCluster, m.Cluster)
		if err != nil {
			return err
		}
	}

	//err = olm.SetupWithManager(m.Manager)
	//if err != nil {
	//	return err
	//}

	if ctrlopts.IsControllerEnabled("tenant", ctrls) {
		err = tenant.SetupWithManager(m.Manager)
		if err != nil {
			return err
		}
	}

	if ctrlopts.IsControllerEnabled("project", ctrls) {
		err = project.SetupWithManager(m.Manager)
		if err != nil {
			return err
		}
	}

	if ctrlopts.IsControllerEnabled("crd", ctrls) {
		err = crds.SetupWithManager(m.Manager, m.PivotClient.Direct())
		if err != nil {
			return err
		}
	}

	if ctrlopts.IsControllerEnabled("quota", ctrls) {
		err = quota.SetupWithManager(m.Manager, m.PivotClient.Direct())
		if err != nil {
			return err
		}
	}

	if ctrlopts.IsControllerEnabled("user", ctrls) {
		err = user.SetupWithManager(m.Manager, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

// setupWithWebhooks set up webhooks into manager
func setupWithWebhooks(m *LocalManager) {
	hookServer := m.GetWebhookServer()
	decoder := admisson.NewDecoder(m.GetScheme())
	hookServer.Register("/warden-validate-tenant-kubeworkz-io-v1-tenant", &webhook.Admission{Handler: tenant2.NewValidator(m.GetClient(), m.IsMemberCluster, decoder)})
	hookServer.Register("/warden-validate-tenant-kubeworkz-io-v1-project", &webhook.Admission{Handler: project2.NewValidator(m.GetClient(), m.IsMemberCluster, decoder)})
	hookServer.Register("/validate-core-kubernetes-v1-resource-quota", &webhook.Admission{Handler: quota2.NewValidator(m.PivotClient.Direct(), m.GetClient(), decoder)})
	hookServer.Register("/warden-validate-hotplug-kubeworkz-io-v1-hotplug", admisson.ValidatingWebhookFor(m.GetScheme(), hotplug2.NewHotplugValidator(m.IsMemberCluster)))
}
