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

package controllers

import (
	"errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/saashqdev/kubeworkz/pkg/clog"
	"github.com/saashqdev/kubeworkz/pkg/ctrlmgr/controllers/binding"
	cluster "github.com/saashqdev/kubeworkz/pkg/ctrlmgr/controllers/cluster"
	"github.com/saashqdev/kubeworkz/pkg/ctrlmgr/controllers/quota"
	"github.com/saashqdev/kubeworkz/pkg/ctrlmgr/options"
	"github.com/saashqdev/kubeworkz/pkg/utils/ctrlopts"
)

var setupFns = make(ctrlopts.ControllerInitFns)

func init() {
	// setup controllers
	setupFns["cluster"] = cluster.SetupWithManager
	setupFns["kuberesourcequota"] = quota.SetupWithManager
	setupFns["clusterrolebinding"] = binding.SetupClusterRoleBindingReconcilerWithManager
	setupFns["rolebinding"] = binding.SetupRoleBindingReconcilerWithManager
}

// SetupWithManager set up controllers into manager
func SetupWithManager(m manager.Manager, controllers string, opts *options.Options) error {
	for name, f := range setupFns {
		if !ctrlopts.IsControllerEnabled(name, ctrlopts.ParseControllers(controllers)) {
			continue
		}
		if err := f(m, opts); err != nil {
			var kindMatchErr *meta.NoKindMatchError
			if errors.As(err, &kindMatchErr) {
				clog.Warn("CRD %v is not installed, its controller will dry run!", kindMatchErr.GroupKind)
				continue
			}
			return err
		}
	}
	return nil
}